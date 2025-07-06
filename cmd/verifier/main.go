package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/axmz/go-distributed-lock/pkg/config"
	redisClient "github.com/axmz/go-distributed-lock/pkg/redis"
	pb "github.com/axmz/go-distributed-lock/proto/report"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type reporterServer struct {
	pb.UnimplementedReporterServer
	mu      sync.Mutex
	results map[string]int
	expect  int
	done    chan struct{}
}

func (s *reporterServer) ReportFinal(ctx context.Context, in *pb.FinalCount) (*pb.Ack, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[in.Id] = int(in.Value)
	log.Printf("Received report from %s: %d", in.Id, in.Value)
	if len(s.results) >= s.expect {
		select {
		case s.done <- struct{}{}:
		default:
		}
	}
	return &pb.Ack{Message: "Received"}, nil
}

func main() {
	cfg := config.Init()
	rc := redisClient.Init(cfg.RedisAddr)
	grpcServer := grpc.NewServer()

	// Init Reporter Server
	rs := &reporterServer{
		results: make(map[string]int),
		expect:  cfg.Replicas,
		done:    make(chan struct{}, 1),
	}
	pb.RegisterReporterServer(grpcServer, rs)

	// Init Health Server
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", cfg.VerifierPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go grpcServer.Serve(lis)
	log.Println("Verifier gRPC server started on ", cfg.VerifierPort)

	<-rs.done
	log.Println("All results received")

	// Get Redis Last Counter Value
	redisVal, err := rc.Get(context.Background(), "counter").Int()
	if err != nil {
		log.Fatalf("Failed to get counter from Redis: %v", err)
	}

	// Find Max Reported Counter Value
	var maxVal int
	for _, v := range rs.results {
		if v > maxVal {
			maxVal = v
		}
	}

	// Calc Expected Counter Value
	expected := cfg.Replicas * cfg.Iterations

	// Compare results
	fmt.Printf("Max reported: %d, Redis counter: %d, Expected counter: %d\n", maxVal, redisVal, expected)

	if maxVal == redisVal && maxVal == expected {
		fmt.Println("SUCCESS: No race conditions detected.")
	} else {
		fmt.Println("FAILURE: Race condition detected!")
	}
}
