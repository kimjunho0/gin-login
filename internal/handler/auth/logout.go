package auth

import (
	"gin-login/middleware"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 로그아웃

// @tags auth
// @Summary logout
// @Description 로그아웃
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Success 200
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/logout [POST]
func Logout(c *gin.Context) {
	//Token 으로부터 ID 얻은거임
	managerId := middleware.GetReqManagerIdFromToken(c.Request)
	//Logout
	session.Logout(managerId)
	c.JSON(http.StatusOK, "로그아웃 완료")
}

type BindRefresh struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
