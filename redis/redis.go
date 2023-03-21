package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Connect() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
}

func Get(key string) (string, bool) {
	exist := true
	val, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		exist = false
	} else if err != nil {
		fmt.Sprintf("redis error :%v", err)
	}
	return val, exist
}
