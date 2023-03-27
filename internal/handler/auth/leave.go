package auth

import (
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @tags auth
// @Summary 회원탈퇴
// @Description 회원 탈퇴
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Param pwd path string true "패스워드"
// @Success 200
// @Failure 400
// @Router /api/auth/leave/{pwd} [DELETE]
func Leave(c *gin.Context) {
	//path 에서 password 받아오기
	pwd, isExist := c.Params.Get("pwd")
	if !isExist {
		panic(cerror.BadRequest())
	}
	//userid 받아오기
	userId := middleware.GetReqManagerIdFromToken(c.Request)

	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	//입력한 패스워드, 토큰에 맞는 아이디
	var user = models.User{
		Id:       userId,
		Password: pwd,
	}
	//기존 pw 가져오기
	var pw string

	if err := tx.Model(&models.User{}).
		Where("id = ?", user.Id).
		Select("password").
		Take(&pw).Error; err != nil {
		panic(cerror.DBErr(err))
	}
	//password 비교
	if !PasswordCompare(pw, user.Password) {
		panic(cerror.BadRequestWithMsg("비번 틀림"))
	}

	if err := tx.Delete(&user).Error; err != nil {
		panic(err)
	}

	tx.Commit()
	//transaction 끝
	c.JSON(http.StatusOK, "유저삭제 완료")
}
