package auth

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterIn struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

func Register(c *gin.Context) {
	var body RegisterIn
	if err := c.ShouldBind(&body); err != nil {
		fmt.Sprintf("bind error", err)
	}
	// 아까 만든 UserDB 에다가 넣을거임
	User := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: middleware.RefreshToken(), //두럭 api 에서는 refresh token 이 바뀌지 않음
		Name:         body.Name,
	}

	//transaction 시작
	tx := migrate.DB.Begin()
	if err := tx.Error; err != nil {
		panic("transaction error")
	}
	defer tx.Rollback()

	if err := tx.Create(&User).Error; err != nil {
		panic("DB create error")
	}

	tx.Commit()
	//transaction 끝

}

func PasswordHash(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic("hash password error")
	}
	return string(hash)
}