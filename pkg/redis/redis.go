package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func Init() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}
