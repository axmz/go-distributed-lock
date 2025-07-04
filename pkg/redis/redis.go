package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func Init(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Retry
	for i := range 10 {
		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			return client
		}
		log.Printf("waiting for redis at %s... (%d/10)", addr, i+1)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("could not connect to redis at %s after 10 attempts", addr)
	return nil
}
