package middleware

import (
	"github.com/google/uuid"
	"strings"
)

func RefreshToken() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
