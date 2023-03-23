package auth

import (
	"fmt"
	"gin-login/migrate"
	"gin-login/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type RegisterIn struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

var mystring = "binding err"

// @tags auth
// @Summary register
// @name register
// @Accept json
// @Produce json
// @Param body body auth.RegisterIn true "전화번호,비밀번호,이름"
// @Success 200 {object} models.User
// @Failure 400
// @Router /api/auth/register [POST]
func Register(c *gin.Context) {
	var body *RegisterIn
	if err := c.ShouldBind(&body); err != nil {
		panic(err)
	}

	// ToDO : 회원가입시
	//- 휴대폰번호 11자리가 아니면 에러반환
	//- 패스워드 정책 준수
	//- 이름에 특수기호 못넣게 들어간다면 에러반환

	// 아까 만든 UserDB 에다가 넣을거임
	user := models.User{
		PhoneNumber:  body.PhoneNumber,
		Password:     PasswordHash(body.Password),
		RefreshToken: RefreshToken(), //두럭 api 에서는 refresh token 이 바뀌지 않음
		Name:         body.Name,
	}

	// TODO : unscoped 로 변경
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

	// TODO : tx 변경
	if Del(body) != nil {
		tx := migrate.DB.Begin()
		if err := tx.Error; err != nil {
			panic("transaction err")
		}
		defer tx.Rollback()
		if err := tx.Create(&user).Error; err != nil {
			panic("register error")
		}
		tx.Commit()
	} else {
		tx := migrate.DB.Begin()
		if err := tx.Error; err != nil {
			panic("transaction err")
		}
		defer tx.Rollback()
		if err := tx.Model(&user).Where("phone_number = ?", body.PhoneNumber).Updates(map[string]interface{}{
			"password":      user.Password,
			"refresh_token": user.RefreshToken,
			"name":          user.Name,
			"deleted_at":    nil,
		}).Error; err != nil {
			panic("update err")
		}
		tx.Commit()
	}

	c.JSON(http.StatusOK, &user)
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
