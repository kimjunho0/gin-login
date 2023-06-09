package auth

import (
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/pkg/cerror/db_error"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

// 회원가입

type RegisterIn struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

// @tags auth
// @Summary register
// @Description 회원가입
// @Accept json
// @Produce json
// @Param body body auth.RegisterIn true "전화번호,비밀번호,이름"
// @Success 200
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/register [POST]
func Register(c *gin.Context) {

	var body *RegisterIn
	if err := c.ShouldBind(&body); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}

	//입력한 폰번호의 길이 확인 & 앞자리 010 인지 확인 <- 이건 나중에 뺄수도
	if len(body.PhoneNumber) < 11 || len(body.PhoneNumber) > 11 || body.PhoneNumber[0:3] != "010" || IfPhoneNumberIncludeChar(body.PhoneNumber) {
		panic(cerror.BadRequestWithMsg(cerror.ErrPhoneNumberReceive))
	}

	// 아까 만든 UserDB 에다가 넣을거임
	user := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: RefreshToken(),
		Name:         body.Name,
	}
	//transaction 시작
	tx := migrate.DB.Begin()
	if err := tx.Error; err != nil {
		panic(cerror.DBErr(err))
	}
	defer tx.Rollback()

	// TODO : 에러 핸들링 dooluck-api cerror 패키지 확인하면서 만들어보기

	// 이름, 비번 규칙 확인
	nameValidity(body.Name)
	PasswordValidity(body.Password, body.PhoneNumber)

	// 여기는 문제가 없음
	// false면 create true면 update
	userExist := ifDeletedUser(tx, body)

	if !userExist {

		//회원가입
		if err := tx.Create(&user).Error; err != nil {
			//duplicate 검사
			if db_error.IsUniqueViolation(err) {
				if strings.Contains(err.Error(), "phone_number") {
					panic(cerror.BadRequestWithMsg("이미 가입된 전화번호 입니다."))
				}
			}
			panic(cerror.DBErr(err))
		}
		//재가입
	} else {
		if err := tx.Model(&user).Unscoped().
			Where("phone_number = ?", body.PhoneNumber).
			Updates(map[string]interface{}{
				"password":      user.Password,
				"refresh_token": user.RefreshToken,
				"name":          user.Name,
				"deleted_at":    nil,
			}).Error; err != nil {
			panic(cerror.DBErr(err))
		}
	}

	tx.Commit()
	//transaction 끝

	c.JSON(http.StatusOK, "회원가입 완료")
}

// refresh token 생성
func RefreshToken() string {
	return strings.Replace(uuid.New().String(), "-", "", -1) // refresh token 의 exp 존재하지 않음
}

// password hash 값으로 변환
func PasswordHash(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(cerror.Unknown(err))
	}
	return string(hash)
}

// 비번,이름 툴
var isStringSpecialChar = regexp.MustCompile(`[\{\}\[\]\/?.,;:|\)*~!^\-_+<>@\#$%&\\\=\(\'\"\n\r]+`).MatchString
var isStringNum = regexp.MustCompile(`[0-9]`).MatchString
var isStringAlphabet = regexp.MustCompile(`[a-zA-Z]`).MatchString

// 이름에 특수문자 안들어가게
func nameValidity(name string) {
	if isStringSpecialChar(name) {
		panic(cerror.BadRequestWithMsg("이름에 특수문자를 포함할 수 없습니다."))
	}
}

const (
	errPasswordLenTpl                = "비밀번호는 8자 이상이어야 합니다."
	errPhoneNumberPasswordEqual      = "전화번호와 비밀번호를 동일하게 설정할 수 없습니다."
	errPasswordShouldContainsAllType = "영문, 숫자, 특수문자를 각각 최소 1개 이상 포함되어야 합니다."
	errPasswordContainsKr            = "비밀번호에 한글을 포함할 수 없습니다."
	errSameNumberRepetition          = "동일한 숫자를 3회 이상 반복할 수 없습니다."
	errSameEngRepetition             = "동일한 문자를 3회 이상 반복할 수 없습니다."
	errContinuousNumber              = "연속된 숫자를 3개 이상 사용할 수 없습니다."
	errContinuousEng                 = "연속된 문자를 3개 이상 사용할 수 없습니다."
)

// 비번 조건 확인
func PasswordValidity(pw string, number string) {
	//비번 길이 확인
	if len(pw) < 8 {
		panic(cerror.BadRequestWithMsg(errPasswordLenTpl))
	}
	//전번,비번 동일한지 확인
	if number == pw {
		panic(cerror.BadRequestWithMsg(errPhoneNumberPasswordEqual))
	}
	//영어,숫자,특수문자 하나씩 들어가게 만들기
	if !isStringNum(pw) || !isStringAlphabet(pw) || !isStringSpecialChar(pw) {
		panic(cerror.BadRequestWithMsg(errPasswordShouldContainsAllType))
	}

	//연속,동일된 문자가 3개이상 되지 않게 (숫자,영어) & 한글 사용 금지
	for index, str := range pw {
		if index < len(pw)-2 {
			if str >= 12593 && str <= 55203 {
				panic(cerror.BadRequestWithMsg(errPasswordContainsKr))
			}
			if unicode.IsDigit(rune(pw[index])) && unicode.IsDigit(rune(pw[index+1])) && unicode.IsDigit(rune(pw[index+2])) {
				//동일숫자
				if pw[index] == pw[index+1] && pw[index] == pw[index+2] {
					panic(cerror.BadRequestWithMsg(errSameNumberRepetition))
				}
				//연속숫자
				if pw[index]+1 == pw[index+1] && pw[index]+2 == pw[index+2] {
					panic(cerror.BadRequestWithMsg(errContinuousNumber))
				}
			}
			//연속으로 문자가 쓰였을 경우 그 문자가 연속된 문자인지 확인
			if unicode.IsLetter(rune(pw[index])) && unicode.IsLetter(rune(pw[index+1])) && unicode.IsLetter(rune(pw[index+2])) {
				//연속 문자
				if pw[index]+1 == pw[index+1] && pw[index]+2 == pw[index+2] {
					panic(cerror.BadRequestWithMsg(errContinuousEng))
				}
				//동일 문자
				if pw[index] == pw[index+1] && pw[index] == pw[index+2] {
					panic(cerror.BadRequestWithMsg(errSameEngRepetition))
				}
			}

		}
	}
}

func ifDeletedUser(tx *gorm.DB, body *RegisterIn) bool {
	var model *models.User
	//err 가 있으면 false 반환, 없으면 다음 조건문 실행
	if err := tx.Unscoped().Where("phone_number = ?", body.PhoneNumber).Find(&model).Error; err != nil {
		// record not found 가 존재하면 false 반환하기 = 레코드가 존재하지 않으니 아직 회원가입을 한적도 없는것
		if db_error.IsRecordNotFound(err) {
			return false //create
		} else {
			panic(cerror.DBErr(err))
		}
	}
	//false 면 deleted_at 이 nil
	if model.DeletedAt.Valid == false {
		return false
	}
	return true

	////다음 조건 deleted at 이 널인지 확인
	//if err := tx.Unscoped().Where("phone_number = ?", body.PhoneNumber).Where("deleted_at IS NOT NULL").Take(&model).Error; err != nil {
	//	//deleted at 이 비워져 있으면 false 반환
	//	if db_error.IsRecordNotFound(err) {
	//		return false
	//	} else {
	//		panic(cerror.DBErr(err))
	//	}
	//}
	//phone number 도 존재하며 deleted_at 이 비어있지 않으면 true 반환

}
