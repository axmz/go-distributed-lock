package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/axmz/go-distributed-lock/pkg/config"
	redisClient "github.com/axmz/go-distributed-lock/pkg/redis"

	pb "github.com/axmz/go-distributed-lock/proto/report"
	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	lockKey    = "my-lock"
	counterKey = "counter"
)

func backoffWithJitter(base time.Duration, maxJitter time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(maxJitter)))
	return base + jitter
}

type cfg struct {
	RedisAddr        string
	RedicClient      *redis.Client
	LockType         string
	Iterations       int
	BackoffBase      time.Duration
	MaxJitter        time.Duration
	SimulateWorkTime time.Duration
	LockExpiry       time.Duration
}

func redisLock(ctx context.Context, id string, cfg cfg) int64 {
	locker := redislock.New(cfg.RedicClient)
	var lastSeen int64

	for range cfg.Iterations {
		var lock *redislock.Lock
		var err error

		for {
			lock, err = locker.Obtain(ctx, lockKey, cfg.LockExpiry, nil)
			if err == redislock.ErrNotObtained {
				time.Sleep(backoffWithJitter(cfg.BackoffBase, cfg.MaxJitter))
				continue
			} else if err != nil {
				log.Fatalf("Could not obtain lock: %v", err)
			}
			break
		}

		val, err := cfg.RedicClient.Incr(ctx, counterKey).Result()
		if err != nil {
			log.Fatalf("Could not increment counter: %v", err)
		}

		time.Sleep(cfg.SimulateWorkTime)

		log.Printf("%v, Incremented counter: %d", id, val)
		lastSeen = val

		lock.Release(ctx)
	}

	return lastSeen
}

func naiveRedisLock(ctx context.Context, id string, cfg cfg) int64 {
	var lastSeen int64

	for range cfg.Iterations {
		// Try to acquire the lock, waiting until it's available
		for {
			ok, err := cfg.RedicClient.SetNX(ctx, lockKey, id, cfg.LockExpiry).Result()
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
		val, err := cfg.RedicClient.Incr(ctx, counterKey).Result()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v, Incremented counter: %d", id, val)

		// Simulate work
		time.Sleep(cfg.SimulateWorkTime)

		lastSeen = val
		// Release the lock only if we still own it
		v, err := cfg.RedicClient.Get(ctx, lockKey).Result()
		if err == nil && v == id {
			log.Printf("%v, Release the lock: %d", id, val)
			cfg.RedicClient.Del(ctx, lockKey)
		}
	}
	return lastSeen
}

func main() {
	var ctx = context.Background()

	var env = config.GetEnv("ENVIRONMENT", "local")
	switch env {
	case "local":
		_ = godotenv.Overload()
	case "k8s":
		_ = godotenv.Load()
	default:
		log.Fatalf("Unknown environment: %s", env)
	}
	var redisAddr = config.GetEnv("REDIS_ADDR", "redis:6379")
	var lockType = config.GetEnv("LOCK_TYPE", "redislock") // "naive" or "redislock"
	var iterations = config.GetEnvInt("ITERATIONS", 1000)
	var backoffBase = config.GetEnvDuration("BACKOFF_BASE", 50*time.Millisecond)
	var maxJitter = config.GetEnvDuration("MAX_JITTER", 100*time.Millisecond)
	var simulateWorkTime = config.GetEnvDuration("SIMULATE_WORK_TIME", 100*time.Millisecond)
	var lockExpiry = config.GetEnvDuration("LOCK_EXPIRY", 5000*time.Microsecond)

	rc := redisClient.Init()

	cfg := cfg{
		RedisAddr:        redisAddr,
		RedicClient:      rc,
		LockType:         lockType,
		Iterations:       iterations,
		BackoffBase:      backoffBase,
		MaxJitter:        maxJitter,
		SimulateWorkTime: simulateWorkTime,
		LockExpiry:       lockExpiry,
	}

	id := uuid.New().String()

	var lastSeen int64
	if lockType == "naive" {
		lastSeen = naiveRedisLock(ctx, id, cfg)
	} else {
		lastSeen = redisLock(ctx, id, cfg)
	}

	conn, err := grpc.NewClient("verifier:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to verifier: %v", err)
	}
	defer conn.Close()
	c := pb.NewReporterClient(conn)
	_, err = c.ReportFinal(ctx, &pb.FinalCount{Id: id, Value: lastSeen})
	if err != nil {
		log.Fatalf("Failed to report final value: %v", err)
	}
	log.Printf("Reported final value %d to verifier", lastSeen)
}
