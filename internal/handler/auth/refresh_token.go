package auth

import (
	"github.com/gin-gonic/gin"
)

type BindRefresh struct {
	RefreshToken string `json:"refresh_token"binding:"required"`
}

// refresh token 만들기

// refresh token 바인딩 해서 바인딩한

func RefreshAccessToken(c *gin.Context) {
	var body BindRefresh
	if err := c.ShouldBind(&body); err != nil {
		panic("create access token binding")
	}

}
