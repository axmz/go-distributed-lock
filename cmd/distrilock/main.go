package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/axmz/go-distributed-lock/pkg/config"
	redisClient "github.com/axmz/go-distributed-lock/pkg/redis"
	"github.com/redis/go-redis/v9"

	pb "github.com/axmz/go-distributed-lock/proto/report"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	lockKey    = "my-lock"
	counterKey = "counter"
)

func backoffWithJitter(base time.Duration, maxJitter time.Duration) time.Duration {
	if maxJitter <= 0 {
		return base
	}
	jitter := time.Duration(rand.Int63n(int64(maxJitter)))
	return base + jitter
}

func naiveRedisLock(ctx context.Context, id string, rc *redis.Client, cfg config.Config) int64 {
	var lastSeen int64

	for range cfg.Iterations {
		// Try to acquire the lock, waiting until it's available
		for {
			ok, err := rc.SetNX(ctx, lockKey, id, cfg.LockExpiry).Result()
			if err != nil {
				log.Fatal(err)
			}
			if ok {
				log.Printf("%v, Acquired the lock", id)
				break // Acquired the lock
			}
			time.Sleep(backoffWithJitter(cfg.BackoffBase, cfg.MaxJitter))
		}

		// Critical section: safely increment the counter
		val, err := rc.Incr(ctx, counterKey).Result()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v, Incremented counter: %d", id, val)

		// Simulate work
		time.Sleep(cfg.SimulateWorkTime)

		lastSeen = val
		// Release the lock only if we still own it
		v, err := rc.Get(ctx, lockKey).Result()
		if err == nil && v == id {
			log.Printf("%v, Release the lock: %d", id, val)
			rc.Del(ctx, lockKey)
		}
	}
	return lastSeen
}

func ConnectAfterVerifierReady(address string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to verifier at %s: %v", address, err)
	}
	healthClient := healthpb.NewHealthClient(conn)

	for range 10 { // TODO: define max attempts
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: define timeouts
		defer cancel()
		resp, healthErr := healthClient.Check(ctx, &healthpb.HealthCheckRequest{})
		if healthErr == nil && resp.GetStatus() == healthpb.HealthCheckResponse_SERVING {
			return conn, nil
		}
		log.Printf("Waiting for verifier at %s...", address)
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("verifier at %s is not ready after 10 attempts", address)
}

func main() {
	var ctx = context.Background()

	cfg := config.Init()

	rc := redisClient.Init(cfg.RedisAddr)

	conn, err := ConnectAfterVerifierReady(cfg.VerifierAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	c := pb.NewReporterClient(conn)

	id := uuid.New().String()

	lastSeen := naiveRedisLock(ctx, id, rc, cfg)

	_, err = c.ReportFinal(ctx, &pb.FinalCount{Id: id, Value: lastSeen})
	if err != nil {
		log.Fatalf("Failed to report final value: %v", err)
	}

	log.Printf("Reported final value %d to verifier", lastSeen)
}
