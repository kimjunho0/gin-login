package auth

import (
	"gin-login/middleware"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Logout(c *gin.Context) {
	//Token 으로부터 ID 얻은거임
	managerId := middleware.GetReqManagerIdFromToken(c.Request)
	//Logout
	session.Logout(managerId)
	c.Status(http.StatusOK)
}
