package auth

import (
	"gin-login/middleware"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @tags auth
// @Summary logout 기능
// @name logout
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Success 200
// @Failure 400
// @Router /api/auth/logout [POST]
func Logout(c *gin.Context) {
	//Token 으로부터 ID 얻은거임
	managerId := middleware.GetReqManagerIdFromToken(c.Request)
	//Logout
	session.Logout(managerId)
	c.Status(http.StatusOK)
}
