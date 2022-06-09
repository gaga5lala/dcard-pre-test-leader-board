package store

import (
	"github.com/go-redis/redis/v9"
)

func NewRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return redisClient
}
