package auth

import (
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
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
// @Success 200
// @Failure 400
// @Router /api/auth/register [POST]
func Register(c *gin.Context) {
	var body *RegisterIn
	if err := c.ShouldBind(&body); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
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
	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	// TODO : unscoped 로 변경

	//deleted at 을 찾는데 못찾으면 err 값 반환

	Del := func(body *RegisterIn) bool {
		model := models.User{
			PhoneNumber: body.PhoneNumber,
		}
		if err := tx.Where("phone_number = ?", body.PhoneNumber).Find(&model); err != nil {
			isLeave := tx.Where("phone_number = ?", body.PhoneNumber).Where("deleted_at IS NOT NULL").Find(&model)
			if isLeave != nil {
				return true
			}
			return false
		}
		return false

	}

	// TODO : tx 변경 --완료--
	// 여기는 문제가 없음
	// err 값이 없으면 update err가 있으면 create
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
