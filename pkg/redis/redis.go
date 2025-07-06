package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func Init(addr string) redis.Cmdable {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{addr},
	})

	// Retry
	for i := range 10 {
		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			return client
		}
		log.Printf("waiting for redis cluster at %s... (%d/10)", addr, i+1)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("could not connect to redis cluster at %s after 10 attempts", addr)
	return nil
}
