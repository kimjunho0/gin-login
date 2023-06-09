package middleware

import (
	"fmt"
	"gin-login/internal/constants"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AuthenticationMiddleware(c *gin.Context) {
	managerId := GetReqManagerIdFromToken(c.Request)
	c.Set(constants.CtxReqManagerId, managerId)
	c.Next()
}

func GetReqManagerId(c *gin.Context) int {
	managerId, exist := c.Get(constants.CtxReqManagerId)
	if !exist {
		cerror.Forbidden()
	}
	return managerId.(int)
}

// = FetchManagerByPhoneNumber
func TakeManagerInformation(phoneNumber string, project ...string) *models.User {
	user := models.User{
		PhoneNumber: phoneNumber,
	}
	if err := migrate.DB.Select(project).Where("phone_number = ?", user.PhoneNumber).Take(&user).Error; err != nil {
		panic(cerror.ForbiddenWithMsg("일치하는 사용자가 없습니다."))
	}
	return &user
}

// =FetchmanagerId
func GetInforUserById(id int, project ...string) *models.User {
	var body = models.User{
		Id: id,
	}
	if err := migrate.DB.Select(project).Take(&body).Error; err != nil {
		panic(cerror.DBErr(err))
	}
	return &body
}

//여기서부턴 jwt

// access token id 에 따라 다르게 만들어짐
func CreatAccessToken(id int) (string, int64) {
	expiresAt := time.Now().Add(10 * time.Minute).Unix()

	claims := &jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Subject:   fmt.Sprintf("%d", id), //sub 자리에 id 를 넣은 access 토큰 생성
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte("SECRET")) // SECRET 코드로 string 으로 바꿈
	if err != nil {
		panic(cerror.Unknown(err))
	}
	return ss, expiresAt
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}
type AccessAndRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    string `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func GetReqManagerIdFromToken(r *http.Request) int {
	token, claims := ParseTokenClaims(r) //access token 값임
	if token.Valid {                     //jwt.token.valid 로 유효성 검사 토큰이 있는지 없는지 검사하는건가? valid 는 bool 타입 반환함
		managerId, _ := strconv.Atoi(claims["sub"].(string))
		//access token create 할때 넣은 id 자리값을 claims["sub"]으로 다시 받아오는듯

		//Session 체크 (access 토큰 생성되어 있어야 )
		valid, reason := session.IsValid(managerId, token.Raw)
		//valid 가 존재하지 않으면 = 유효기간 만료
		if !valid {
			if reason == session.Expired {
				panic(cerror.Forbidden())
			} else if reason == session.MultiLogin {
				panic(cerror.ForbiddenWithMsg(cerror.ErrMultiLogin))
			}
		}
		return managerId

	} else {
		panic(cerror.Forbidden())
	}

}

// Valid 검사 유무만 다름

func GetReqManagerIdWithoutExpValidationFromToken(r *http.Request) int {
	_, claims := ParseTokenClaimsWithoutExpValidation(r)
	managerId, _ := strconv.Atoi(claims["sub"].(string))
	return managerId
}

// 일단은 로그인 할때 나오는 정보인듯 (AccessToken and RefreshToken and ExpiresAt
func MakeAccessAndRefreshResponse(accessToken string, expiresAt int64, refreshToken string) *AccessAndRefreshResponse {
	h, m, s := time.Unix(expiresAt, 0).Clock()
	return &AccessAndRefreshResponse{
		AccessToken:  accessToken,
		ExpiresAt:    fmt.Sprintf("로그인 유효 시간 %d시%d분%d초", h, m, s),
		RefreshToken: refreshToken,
	}
}

// bearer Token 에서 claims 정보 가져오기     토큰하고     토큰 claims 반환(정보)
func ParseTokenClaims(r *http.Request) (*jwt.Token, jwt.MapClaims) {
	tokenString := ParseBearerToken(r) //access토큰 값 auth-token 에서 가져온거임
	//token 파싱
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unepected signing method %v", token.Header["alg"])
		}
		return []byte("SECRET"), nil
	})
	if err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok { //token.Claims 로 파싱한 토큰값 정보를 가져온거같음
		return token, claims //토큰값과 토큰정보 반환
	} else {
		panic(cerror.Forbidden())
	}
}
func ParseTokenClaimsWithoutExpValidation(r *http.Request) (*jwt.Token, jwt.MapClaims) {
	tokenString := ParseBearerToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unepected signing method %v", token.Header["alg"])
		}
		return []byte("SECRET"), nil
	})
	if err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return token, claims
	} else {
		panic(cerror.Forbidden())
	}
}

// request.HeaderExtractor 에서 내가보기엔 auth-token 부분을 ExtractToken으로 가져온듯 함
// access token에서 가져온걸까?
// auth-token 이라는 헤더값에서 토큰을 가져오는거임 access token 값이 들어가는것도 맞음

func ParseBearerToken(r *http.Request) string {
	token, err := request.HeaderExtractor([]string{"auth-token"}).ExtractToken(r)
	if err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}

	return strings.TrimPrefix(token, "Bearer ")
}
