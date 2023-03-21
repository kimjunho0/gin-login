package redis

import (
	"context"
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
		panic("redis get error")
	}
	return val, exist
}

func Set(key string, value string, ttl time.Duration) {
	err := client.Set(context.Background(), key, value, ttl).Err()
	if err != nil {
		panic("redis set error")
	}
}

func Delete(key string) {
	err := client.Del(context.Background(), key).Err()
	if err != nil {
		panic("redis del error")
	}
}
