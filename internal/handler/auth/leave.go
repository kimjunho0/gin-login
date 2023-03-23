package auth

import (
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Deleted_User struct {
	Password string `json:"password" binding:"required"`
}

// @tags auth
// @Summary 회원탈퇴
// @Description 회원 탈퇴
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Param body body auth.Deleted_User true "비밀번호"
// @Success 200
// @Failure 400
// @Router /api/auth/leave [POST]
func Leave(c *gin.Context) {
	var body Deleted_User
	if err := c.ShouldBind(&body); err != nil {
		panic("Leave binding error")
	}
	//지금 로그인 한 유저 아이디(ID)
	userId := middleware.GetReqManagerIdFromToken(c.Request)
	//입력한 패스워드
	var user = models.User{
		Id:       userId,
		Password: body.Password,
	}
	//기존 pw 가져오기
	var pw string

	migrate.DB.Model(&models.User{}).Where("id = ?", user.Id).Select("password").Take(&pw)

	if !PasswordCompare(pw, user.Password) {
		panic("비밀번호가 틀렸습니다.")
	}
	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()
	migrate.DB.Model(&user).Where("id = ?", user.Id).Update("deleted_at", time.Now())
	tx.Commit()
	//끝
	c.Status(http.StatusOK)
}
