package auth

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
// @Param auth-token header string true "access token"
// @Param body body auth.BindRefresh true "갱신"
// @Success 200 {object} middleware.AccessTokenResponse
// @Failure 400
// @Router /api/auth/refresh-token [POST]
func RefreshAccessToken(c *gin.Context) {
	var body BindRefresh
	if err := c.ShouldBind(&body); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}

	userId := middleware.GetReqManagerIdWithoutExpValidationFromToken(c.Request)
	userRefresh := middleware.GetInforUserById(userId, "refresh_token") //refresh token userid 로 받아오기

	//입력한 refresh 값과 db의 refresh 값이 다르면 인증정보 만료 반환
	if body.RefreshToken != userRefresh.RefreshToken {
		panic(cerror.BadRequestWithMsg(cerror.ErrRefreshTokenInvalid))
	}

	token, expiresAt := middleware.CreatAccessToken(userId)
	//새로운 토큰으로 세션 로그인

	session.Login(userId, token, AccessTokenTimeOut)

	// token, expire 반환 expire = 분단위로 반환
	h, m, s := time.Unix(expiresAt, 0).Clock()
	c.JSON(http.StatusOK, middleware.AccessTokenResponse{AccessToken: token,
		ExpiresAt: fmt.Sprintf("로그인 유효시간 %d시%d분%d초",
			h, m, s)})

}
