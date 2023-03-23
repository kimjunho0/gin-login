package redis

import (
	"context"
	"gin-login/pkg/cerror"
	"github.com/go-redis/redis/v8"
	"time"
)

var client *redis.Client

func Connect() {
	client = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   1,
	})
}

func Get(key string) (string, bool) {
	exist := true
	val, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		exist = false
	} else if err != nil {
		panic(cerror.RedisErr(err))
	}
	return val, exist
}

// ex s_1, access token 만든거, 10분 이런식으로 들어가서
// key 는 s_1, value는 access token , 유효기간은 10분 으로 key set 을 하는듯
func Set(key string, value string, ttl time.Duration) {
	err := client.Set(context.Background(), key, value, ttl).Err()
	if err != nil {
		panic(cerror.RedisErr(err))
	}
}

func Delete(key string) {
	err := client.Del(context.Background(), key).Err()
	if err != nil {
		panic(cerror.RedisErr(err))
	}
}
