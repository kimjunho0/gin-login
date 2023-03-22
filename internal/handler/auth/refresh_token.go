package auth

import (
	"gin-login/middleware"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BindRefresh struct {
	RefreshToken string `json:"refresh_token"binding:"required"`
}

// refresh token 만들기

// refresh token 바인딩
// @tags auth
// @Summary  refresh token
// @Description refresh token 으로 access token 갱신
// @name refresh token
// @Accept json
// @Produce json
// @Param auth-token header string true "$access token"
// @Param body body auth.BindRefresh true "갱신"
// @Success 200 {object} middleware.AccessTokenResponse
// @Failure 400
// @Router /api/auth/refresh-token [POST]
func RefreshAccessToken(c *gin.Context) {
	var body BindRefresh
	if err := c.ShouldBind(&body); err != nil {
		panic("create access token binding")
	}
	userId := middleware.GetReqManagerIdWithoutExpValidationFromToken(c.Request)
	userRefresh := middleware.GetInforUserById(userId, "refresh_token") //refresh token userid 로 받아오기
	if body.RefreshToken != userRefresh.RefreshToken {
		panic("refresh token 이 일치하지 않음")
	}
	token, expiresAt := middleware.CreatAccessToken(userId)
	//새로운 토큰으로 세션 로그인

	session.Login(userId, token, AccessTokenTimeOut)

	c.JSON(http.StatusOK, middleware.AccessTokenResponse{AccessToken: token, ExpiresAt: expiresAt})

}
