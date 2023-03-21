package auth

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Needlogin struct {
	phoneNumber string `json:"phone_number" binding:"required"`
	password    string `json:"password" binding:"required"`
}

const AccessTokenTimeOut = 10 * time.Minute

func Login(c *gin.Context) {
	var login Needlogin
	if err := c.ShouldBindJSON(&login); err != nil {
		fmt.Sprintf("login binding error", err)
	}
	//manager.go 의 phonenumber 에 맞는 user 구조체 가져오기
	fmt.Printf("%v", login)

	//입력한 폰번호와 DB에 있는 폰번호가 일치하는지 확인, 있으면 가져옴
	manager := middleware.TakeManagerInformation(login.phoneNumber, "id", "password", "refresh_token", "num_password_fail")

	if manager.NumPasswordFail >= 10 {
		panic("비밀번호 10회 오류")
	}

	if !PasswordCompare(manager.Password, login.password) {
		//비밀번호 불일치
		if err := migrate.DB.Model(&manager).
			Where("phone_number = ?", login.phoneNumber).
			Update("num_password_fail", gorm.Expr("num_password_fail + 1")).Error; err != nil {
			panic("db error")
		}
		if manager.NumPasswordFail+1 >= 10 {
			panic("비밀번호 10회 오류입니다. 서비스를 이용하시려면 비밀번호를 변경해주세요")
		} else {
			panic(fmt.Sprintf("비밀번호 %d 회 오류입니다 10회 초과시 서비스 이용이 제한됩니다.", manager.NumPasswordFail+1))
		}
	}

	//비번 일치
	if err := migrate.DB.Model(&models.User{}).
		Where("phone_number = ?", manager.PhoneNumber).
		Update("num_password_fail", 0).Error; err != nil {
		panic(fmt.Sprintf("DB 오류 %v", err))
	}

	//access 토큰 생성
	accessToken, expiresAt := middleware.CreatAccessToken(manager.Id)
	//manager 구조체 가져온걸로 계속 활용
	session.Login(manager.Id, accessToken, AccessTokenTimeOut)

	c.JSON(http.StatusOK, middleware.MakeAccessAndRefreshResponse(accessToken, expiresAt, manager.RefreshToken))
}

// 패스워드 일치 확인
func PasswordCompare(hashPw string, plainPw string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPw), []byte(plainPw)); err != nil {
		return false
	}
	return true
}
