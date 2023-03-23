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

// @tags auth
// @Summary 로그인정보
// @Description 로그인한 자기 정보 가져오기
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Success 200 {object} auth.GetInfo
// @Failure 400
// @Router /api/auth/info [GET]
func Info(c *gin.Context) {
	userId := middleware.GetReqManagerIdFromToken(c.Request)
	userInfo := middleware.GetInforUserById(userId, "phone_number", "name")

	c.JSON(http.StatusOK, userInfo)
}
