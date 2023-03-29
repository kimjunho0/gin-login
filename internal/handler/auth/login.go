package auth

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"time"
)

type NeedLogin struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

var IfPhoneNumberIncludeChar = regexp.MustCompile(`[a-zA-Zㄱ-힣]`).MatchString

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

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       fmt.Sprintf("%d", manager.Id),
			Username: login.PhoneNumber,
		})
	})
	c.JSON(http.StatusOK, middleware.MakeAccessAndRefreshResponse(accessToken, expiresAt, manager.RefreshToken))
}

// db에 있는 패스워드와 입력받은 패스워드 일치 확인
func PasswordCompare(hashPw string, plainPw string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPw), []byte(plainPw)); err != nil {
		return false
	}
	return true
}
