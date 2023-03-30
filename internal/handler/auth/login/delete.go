package login

import (
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @tags auth
// @Summary delete_user
// @Description 회원 탈퇴
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Success 200
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/delete [DELETE]
func Delete(c *gin.Context) {

	// TODO : 회원가입 로직 다시 생각해보기
	//token 에서 아이디 불러와서
	userId := middleware.GetReqManagerIdFromToken(c.Request)
	// 그 아이디 쿼리 삭제
	if err := migrate.DB.Delete(&models.User{
		Id: userId,
	}).Error; err != nil {
		cerror.DBErr(err)
	}
	//session 로그아웃
	session.Logout(userId)
	c.JSON(http.StatusOK, "회원 탈퇴 완료")
}
