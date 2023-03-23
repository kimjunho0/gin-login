package auth

import (
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"github.com/gin-gonic/gin"
	"net/http"
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
		panic(cerror.BadRequestWithMsg(err.Error()))
	}

	userId := middleware.GetReqManagerIdFromToken(c.Request)

	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	//입력한 패스워드
	var user = models.User{
		Id:       userId,
		Password: body.Password,
	}
	//기존 pw 가져오기
	var pw string

	if err := tx.Model(&models.User{}).
		Where("id = ?", user.Id).
		Select("password").
		Take(&pw).Error; err != nil {
		panic(cerror.DBErr(err))
	}

	if !PasswordCompare(pw, user.Password) {
		c.JSON(http.StatusBadRequest, "비밀번호가 틀렸습니다.")
		panic(cerror.BadRequest())
	}

	if err := tx.Delete(&user).Error; err != nil {
		panic(err)
	}

	tx.Commit()
	//끝
	c.JSON(http.StatusOK, "유저삭제 완료")
}
