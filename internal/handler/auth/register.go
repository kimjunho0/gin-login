package auth

import (
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

type RegisterIn struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

var mystring = "binding err"

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
	if len(body.PhoneNumber) < 11 || len(body.PhoneNumber) > 11 || body.PhoneNumber[0:3] != "010" {
		panic(cerror.BadRequestWithMsg(cerror.ErrPhoneNumberReceive))
	}
	// 이름, 비번 규칙 확인
	NameValidity(body.Name)
	PasswordValidity(body.Password, body.PhoneNumber)
	// ToDO : 회원가입시 --완료--
	//- 휴대폰번호 11자리가 아니면 에러반환 -- 완료 --
	//- 패스워드 정책 준수 --완료--
	//- 이름에 특수기호 못넣게 들어간다면 에러반환 -- 완료 --

	// 아까 만든 UserDB 에다가 넣을거임
	user := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: RefreshToken(), //두럭 api 에서는 refresh token 이 바뀌지 않음
		Name:         body.Name,
	}
	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	// TODO : unscoped 로 변경 --완료..?--

	//deleted at 을 찾는데 못찾으면 err 값 반환

	Del := func(body *RegisterIn) bool {
		model := models.User{
			PhoneNumber: body.PhoneNumber,
		}
		//먼저 전화번호가 데이터베이스에 deleted_at 상관없이 있는지 확인 (unscoped로 deleted가 null이던 null 이 아니던 다 조회)
		if err := tx.Unscoped().Where("phone_number = ?", body.PhoneNumber).Take(&model).Error; err == nil {
			//deleted at 이 비어있지 않으면 true 반환
			err := tx.Unscoped().Where("phone_number = ?", body.PhoneNumber).Where("deleted_at IS NOT NULL").Take(&model).Error
			if err == nil {
				return true
			}
			c.JSON(http.StatusBadRequest, "이미 가입된 전화번호입니다.")
			panic(cerror.BadRequestWithMsg("이미 가입된 전화번호입니다."))
		}
		return false

	}

	// TODO : tx 변경 --완료--
	// 여기는 문제가 없음
	// false면 create true면 update
	if !Del(body) {
		if err := tx.Error; err != nil {
			panic(cerror.DBErr(err))
		}
		if err := tx.Create(&user).Error; err != nil {
			panic(cerror.DBErr(err))
		}
	} else {
		if err := tx.Error; err != nil {
			panic(cerror.DBErr(err))
		}
		if err := tx.Model(&user).Unscoped().Where("phone_number = ?", body.PhoneNumber).Updates(map[string]interface{}{
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

	c.JSON(http.StatusOK, "회원가입 완로")
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
func NameValidity(name string) {
	if isStringSpecialChar(name) {
		panic(cerror.BadRequestWithMsg("이름에 특수문자를 포함할 수 없습니다."))
	}
}

// 비번 조건 확인
func PasswordValidity(pw string, number string) {
	if len(pw) < 8 {
		panic(cerror.BadRequestWithMsg("비밀번호는 8자 이상이어야 합니다."))
	}
	if number == pw {
		panic(cerror.BadRequestWithMsg("전화번호와 비밀번호를 동일하게 설정할 수 없습니다"))
	}
	//영어,숫자,특수문자 하나씩 들어가게 만들기
	if !isStringNum(pw) || !isStringAlphabet(pw) || !isStringSpecialChar(pw) {
		panic(cerror.BadRequestWithMsg("영문,숫자,특수문자가 각각 하나 이상 포함되어야 합니다."))
	}
	//연속,동일된 문자가 3개이상 되지 않게 (숫자,영어) & 한글 사용 금지
	for index, str := range pw {
		if str >= 12593 && str <= 55203 {
			panic(cerror.BadRequestWithMsg("비밀번호에 한글을 포함할 수 없습니다."))
		}
		if unicode.IsDigit(rune(pw[index])) && unicode.IsDigit(rune(pw[index+1])) && unicode.IsDigit(rune(pw[index+2])) {
			//동일숫자
			if pw[index] == pw[index+1] && pw[index] == pw[index+2] {
				panic(cerror.BadRequestWithMsg("동일한 숫자를 3회 이상 반복할 수 없습니다."))
			}
			//연속숫자
			if pw[index]+1 == pw[index+1] && pw[index]+2 == pw[index+2] {
				panic("연속된 숫자를 3회 이상 사용할 수 없습니다.")
			}
		}
		//연속으로 문자가 쓰였을 경우 그 문자가 연속된 문자인지 확인
		if unicode.IsLetter(rune(pw[index])) && unicode.IsLetter(rune(pw[index+1])) && unicode.IsLetter(rune(pw[index+2])) {
			//연속 문자
			if pw[index]+1 == pw[index+1] && pw[index]+2 == pw[index+2] {
				panic("연속된 문자를 3회 이상 사용할 수 없습니다.")
			}
			//동일 문자
			if pw[index] == pw[index+1] && pw[index] == pw[index+2] {
				panic("동일한 문자를 3회 이상 반복할 수 없습니다.")
			}
		}

	}

}
