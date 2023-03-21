package session

import (
	"fmt"
	"gin-login/redis"
	"time"
)

// 로그인 유무
func sessionKey(userId int) string {
	return fmt.Sprintf("%s_%d", "s", userId)
}

func Login(userid int, token string, ttl time.Duration) {
	redis.Set(sessionKey(userid), token, ttl)
}

func Logout(userid int) {
	redis.Delete(sessionKey(userid))
}
