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

type InvalidReason int

const (
	Valid InvalidReason = iota
	Expired
	MultiLogin
)

// 중복 로그인 , 로그인 유효기간, 로그인 중인지 확인 하는듯
// session key 가 client.set 이 되어 있는지 확인하는 과정
func IsValid(userid int, token string) (bool, InvalidReason) {
	val, exist := redis.Get(sessionKey(userid))
	if !exist {
		return false, Expired
	} else if val != token {
		return false, MultiLogin
	}
	return true, Valid

}
