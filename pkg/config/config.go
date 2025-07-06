package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr        string
	VerifierPort     string
	VerifierAddr     string
	Iterations       int
	Replicas         int
	BackoffBase      time.Duration
	MaxJitter        time.Duration
	SimulateWorkTime time.Duration
	LockExpiry       time.Duration
}

func Init() Config {
	_ = godotenv.Overload()

	return Config{
		RedisAddr:        GetEnv("REDIS_ADDR", "redis:6379"),
		VerifierPort:     GetEnv("VERIFIER_PORT", ":50051"),
		VerifierAddr:     GetEnv("VERIFIER_ADDR", "verifier:50051"),
		Iterations:       GetEnvInt("ITERATIONS", 1000),
		Replicas:         GetEnvInt("REPLICAS", 5),
		BackoffBase:      GetEnvDuration("BACKOFF_BASE", 50*time.Millisecond),
		MaxJitter:        GetEnvDuration("MAX_JITTER", 100*time.Millisecond),
		SimulateWorkTime: GetEnvDuration("SIMULATE_WORK_TIME", 100*time.Millisecond),
		LockExpiry:       GetEnvDuration("LOCK_EXPIRY", 5000*time.Microsecond),
	}
}

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func GetEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func GetEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
