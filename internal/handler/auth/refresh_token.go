package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
)

type BindRefresh struct {
	RefreshToken string `json:"refresh_token"binding:"required"`
}

// refresh token 만들기
func RefreshToken() string {
	return strings.Replace(uuid.New().String(), "-", "", -1) // refresh token 의 exp 존재하지 않음
}

func CreateAccessToken(c *gin.Context) {
	var body BindRefresh
	if err := c.ShouldBind(&body); err != nil {
		panic("create access token binding")
	}

}
