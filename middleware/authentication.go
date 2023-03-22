package middleware

import (
	"fmt"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/redis/session"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// = FetchManagerByPhoneNumber
func TakeManagerInformation(phoneNumber string, project ...string) *models.User {
	user := models.User{
		PhoneNumber: phoneNumber,
	}
	if err := migrate.DB.Select(project).Where("phone_number = ?", user.PhoneNumber).Take(&user).Error; err != nil {
		panic("회원 정보가 없습니다.")
	}
	return &user
}

// =FetchmanagerId
func GetInforUserById(id int, project ...string) *models.User {
	var body = models.User{
		Id: id,
	}
	if err := migrate.DB.Select(project).Take(&body).Error; err != nil {
		panic("User by id error")
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
		panic("Unknown")
	}
	return ss, expiresAt
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}
type AccessAndRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func GetReqManagerIdFromToken(r *http.Request) int {
	token, claims := ParseTokenClaims(r)
	if token.Valid { //jwt.token 타입의 valid 로 유효성 검사 토큰이 있는지 없는지 검사하는건가?
		managerId, _ := strconv.Atoi(claims["sub"].(string))
		//access token create 할때 넣은 id 자리값을 claims["sub"]으로 다시 받아오는듯

		//Session 체크 (access 토큰 생성되어 있어야 )
		valid, reason := session.IsValid(managerId, token.Raw)
		//valid 가 존재하지 않으면 = 유효기간 만료지
		if !valid {
			if reason == session.Expired {
				panic("Expired")
			} else if reason == session.MultiLogin {
				panic("Multilogin")
			}
		}
		return managerId

	} else {
		panic("error")
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
	return &AccessAndRefreshResponse{
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
	}
}

// Token 에서 claims 정보 가져오기              토큰하고     토큰 claims 반환(정보)
func ParseTokenClaims(r *http.Request) (*jwt.Token, jwt.MapClaims) {
	tokenString := ParseBearerToken(r) //토큰 값 가져온거임
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unepected signing method %v", token.Header["alg"])
		}
		return []byte("SECRET"), nil
	})
	if err != nil {
		panic(fmt.Sprintf("parse token claims error ", err))
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok { //token.Claims로 토큰 정보를 맵으로 해서 가져온거같음
		return token, claims
	} else {
		panic("왜난거지")
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
		panic(fmt.Sprintf("parse token claims error ", err))
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return token, claims
	} else {
		panic("왜난거지")
	}
}

// 이건 대체 뭘까 하..
// request.HeaderExtractor 에서 내가보기엔 auth-token 부분을 ExtractToken으로 가져온듯 함
// access token에서 가져온걸까?

func ParseBearerToken(r *http.Request) string {
	token, err := request.HeaderExtractor([]string{"auth-token"}).ExtractToken(r)
	if err != nil {
		panic(fmt.Sprintf("ParseBearer error %v", err))
	}
	return strings.TrimPrefix(token, "Bearer ")
}
