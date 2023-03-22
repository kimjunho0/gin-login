package auth

import (
	"fmt"
	"gin-login/migrate"
	"gin-login/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type RegisterIn struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

func Register(c *gin.Context) {
	var body *RegisterIn
	if err := c.ShouldBind(&body); err != nil {
		fmt.Sprintf("bind error", err)
	}
	// 아까 만든 UserDB 에다가 넣을거임
	User := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: RefreshToken(), //두럭 api 에서는 refresh token 이 바뀌지 않음
		Name:         body.Name,
	}
	Del := func(body *RegisterIn) error {
		user := models.User{
			PhoneNumber: body.PhoneNumber,
		}
		if err := migrate.DB.Select("deleted_at").Where("phone_number = ?", user.PhoneNumber).Take(&user).Error; err != nil {
			return err
		}
		return nil
	}
	fmt.Println(Del(body))
	if Del(body) != nil {
		tx := migrate.DB.Begin()
		if err := tx.Error; err != nil {
			panic("transaction err")
		}
		defer tx.Rollback()
		if err := tx.Create(&User).Error; err != nil {
			panic("register error")
		}
		tx.Commit()
	} else {
		tx := migrate.DB.Begin()
		if err := tx.Error; err != nil {
			panic("transaction err")
		}
		defer tx.Rollback()
		if err := tx.Model(&User).Where("phone_number = ?", body.PhoneNumber).Updates(map[string]interface{}{
			"password":      User.Password,
			"refresh_token": User.RefreshToken,
			"name":          User.Name,
			"deleted_at":    nil,
		}).Error; err != nil {
			panic("update err")
		}
		tx.Commit()
	}
}

func RefreshToken() string {
	return strings.Replace(uuid.New().String(), "-", "", -1) // refresh token 의 exp 존재하지 않음
}

func PasswordHash(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic("hash password error")
	}
	return string(hash)
}
