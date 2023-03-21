package middleware

import (
	"fmt"
	"gin-login/migrate"
	"gin-login/models"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// = FetchManagerByPhoneNumber
func TakeManagerInformation(phoneNumber string, project ...string) *models.User {
	user := models.User{
		PhoneNumber: phoneNumber,
	}
	if err := migrate.DB.Select(project).Where("phone_number = ?", user.PhoneNumber).Take(&user).Error; err != nil {
		panic("failed take user inform")
	}
	return &user
}

func CreatAccessToken(id int) (string, int64) {
	expiresAt := time.Now().Add(10 * time.Minute).Unix()

	claims := &jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Subject:   fmt.Sprintf("%d", id),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte("SECRET"))
	if err != nil {
		panic("Unknown")
	}
	return ss, expiresAt
}

type AccessAndRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func MakeAccessAndRefreshResponse(accessToken string, expiresAt int64, refreshToken string) *AccessAndRefreshResponse {
	return &AccessAndRefreshResponse{
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
	}
}
