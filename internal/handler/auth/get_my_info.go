package auth

import (
	"gin-login/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetInfo struct {
	UserId      int    `json:"user_id"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
}

func Info(c *gin.Context) {
	userId := middleware.GetReqManagerIdFromToken(c.Request)
	userinfo := middleware.GetInforUserById(userId, "phone_number", "name")
	a := GetInfo{
		UserId:      userId,
		PhoneNumber: userinfo.PhoneNumber,
		Name:        userinfo.Name,
	}
	c.JSON(http.StatusOK, a)
}
