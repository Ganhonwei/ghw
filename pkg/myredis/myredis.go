package myredis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

func InitRedis(addr string, db int) {
	client = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

func Redis() *redis.Client {
	if client == nil {
		panic("redis client not initialized")
	}
	return client
}
