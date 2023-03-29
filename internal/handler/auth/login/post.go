package login

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var IfPhoneNumberIncludeChar = regexp.MustCompile(`[a-zA-Zㄱ-힣]`).MatchString

// @tags auth
// @Summary logout
// @Description 로그아웃
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Success 200
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/logout [POST]
func Logout(c *gin.Context) {
	//Token 으로부터 ID 얻은거임
	managerId := middleware.GetReqManagerIdFromToken(c.Request)
	//Logout
	session.Logout(managerId)
	c.JSON(http.StatusOK, "로그아웃 완료")
}

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

	// TODO : 회원가입시
	//- 휴대폰번호 11자리가 아니면 에러반환
	//- 패스워드 정책 준수
	//- 이름에 특수기호 못넣게 들어간다면 에러반환

	// 아까 만든 UserDB 에다가 넣을거임
	user := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: RefreshToken(),
		Name:         body.Name,
	}
	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	// TODO : unscoped 로 변경

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
			panic(cerror.BadRequestWithMsg("이미 가입된 전화번호입니다."))
		}
		return false

	}

	// 이름, 비번 규칙 확인
	nameValidity(body.Name)
	PasswordValidity(body.Password, body.PhoneNumber)

	// TODO : tx 변경
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

type BindRefresh struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// refresh token 만들기

// refresh token 바인딩
// @tags auth
// @Summary  refresh token
// @Description refresh token 으로 access token 갱신
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Param body body auth.BindRefresh true "갱신"
// @Success 200 {object} middleware.AccessTokenResponse
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/refresh-token [POST]
func RefreshAccessToken(c *gin.Context) {

	var body BindRefresh
	if err := c.ShouldBind(&body); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}

	userId := middleware.GetReqManagerIdWithoutExpValidationFromToken(c.Request)
	userRefresh := middleware.GetInforUserById(userId, "refresh_token") //refresh token userid 로 받아오기

	//입력한 refresh 값과 db의 refresh 값이 다르면 인증정보 만료 반환
	if body.RefreshToken != userRefresh.RefreshToken {
		panic(cerror.BadRequestWithMsg(cerror.ErrRefreshTokenInvalid))
	}

	token, expiresAt := middleware.CreatAccessToken(userId)
	//새로운 토큰으로 세션 로그인

	session.Login(userId, token, AccessTokenTimeOut)

	// token, expire 반환 expire = 분단위로 반환
	h, m, s := time.Unix(expiresAt, 0).Clock()
	c.JSON(http.StatusOK, middleware.AccessTokenResponse{AccessToken: token,
		ExpiresAt: fmt.Sprintf("로그인 유효시간 %d시%d분%d초",
			h, m, s)})

}

type NeedLogin struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

const (
	maxNumPasswordFailed = 10
	AccessTokenTimeOut   = 10 * time.Minute
)

const (
	errNumPasswordFalExceedTpl = "비밀번호 %d회 오류입니다. 서비스를 이용하시려면 비밀번호를 변경해주세요."
	errPasswordNotMatched      = "비밀번호 %d회 오류입니다. %d회 초과시 서비스 이용이 제한됩니다."
)

// @tags auth
// @Summary login
// @Description 로그인
// @Accept json
// @Produce json
// @Param body body auth.NeedLogin true "전화번호, 비밀번호"
// @Success 200 {object} middleware.AccessAndRefreshResponse
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/login [POST]
func Login(c *gin.Context) {

	var login NeedLogin
	if err := c.ShouldBind(&login); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}
	//입력한 폰번호의 길이 확인
	if len(login.PhoneNumber) < 11 || len(login.PhoneNumber) > 11 || IfPhoneNumberIncludeChar(login.PhoneNumber) {
		panic(cerror.BadRequestWithMsg(cerror.ErrPhoneNumberReceive))
	}

	//user.go 의 phoneNumber 에 맞는 user 구조체 가져오기

	//입력한 폰번호와 DB에 있는 폰번호가 일치하는지 확인, 있으면 가져옴
	manager := middleware.TakeManagerInformation(login.PhoneNumber, "id", "password", "refresh_token", "num_password_fail")

	if manager.NumPasswordFail >= maxNumPasswordFailed {
		panic(cerror.BadRequestWithMsg(fmt.Sprintf(errNumPasswordFalExceedTpl, maxNumPasswordFailed)))
	}

	if !PasswordCompare(manager.Password, login.Password) {
		//비밀번호 불일치
		if err := migrate.DB.Model(&manager).
			Where("phone_number = ?", login.PhoneNumber).
			Update("num_password_fail", gorm.Expr("num_password_fail + 1")).Error; err != nil {
			panic(cerror.DBErr(err))
		}
		if manager.NumPasswordFail+1 >= maxNumPasswordFailed {
			panic(cerror.BadRequestWithMsg(fmt.Sprintf(errNumPasswordFalExceedTpl, maxNumPasswordFailed)))
		} else {
			panic(cerror.BadRequestWithMsg(fmt.Sprintf(errPasswordNotMatched, manager.NumPasswordFail+1, maxNumPasswordFailed)))
		}
	}

	//비번 일치
	if err := migrate.DB.Model(&models.User{}).
		Where("phone_number = ?", manager.PhoneNumber).
		Update("num_password_fail", 0).Error; err != nil {
		panic(cerror.DBErr(err))
	}

	//access 토큰 생성
	accessToken, expiresAt := middleware.CreatAccessToken(manager.Id)

	//session 접속 (redis.set)
	session.Login(manager.Id, accessToken, AccessTokenTimeOut)
	c.JSON(http.StatusOK, middleware.MakeAccessAndRefreshResponse(accessToken, expiresAt, manager.RefreshToken))
}

// db에 있는 패스워드와 입력받은 패스워드 일치 확인
func PasswordCompare(hashPw string, plainPw string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPw), []byte(plainPw)); err != nil {
		return false
	}
	return true
}
