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

// Access token 갱신

// refresh token 바인딩
// @tags auth
// @Summary  refresh token
// @Description refresh token 으로 access token 갱신
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Param body body auth.BindRefresh true "갱신"
// @Success 200 {object} middleware.AccessTokenResponse
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
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

type NeedLogin struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

const (
	maxNumPasswordFailed = 10
	AccessTokenTimeOut   = 10 * time.Minute
)

const (
	errNumPasswordFalExceedTpl = "비밀번호 %d회 오류입니다. 서비스를 이용하시려면 비밀번호를 변경해주세요."
	errPasswordNotMatched      = "비밀번호 %d회 오류입니다. %d회 초과시 서비스 이용이 제한됩니다."
)
