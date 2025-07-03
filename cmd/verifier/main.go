package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	redisClient "github.com/axmz/go-distributed-lock/pkg/redis"
	pb "github.com/axmz/go-distributed-lock/proto/report"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedReporterServer
	mu      sync.Mutex
	results map[string]int64
	expect  int
	done    chan struct{}
}

func (s *server) ReportFinal(ctx context.Context, in *pb.FinalCount) (*pb.Ack, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[in.Id] = in.Value
	if len(s.results) >= s.expect {
		select {
		case s.done <- struct{}{}:
		default:
		}
	}
	return &pb.Ack{Message: "Received"}, nil
}

func main() {
	expect := 5 // TODO: get from env
	s := &server{
		results: make(map[string]int64),
		expect:  expect,
		done:    make(chan struct{}, 1),
	}

	lis, err := net.Listen("tcp", ":50051") // gRPC server port
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterReporterServer(grpcServer, s)
	go grpcServer.Serve(lis)
	log.Println("Verifier gRPC server started on :50051")

	// Wait for all results or timeout
	select {
	case <-s.done:
		log.Println("All results received")
		// TODO: start timeout countdown after first result, then implement a heartbeat mechanism
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for results")
	}

	// Compare with Redis
	rc := redisClient.Init()
	redisVal, err := rc.Get(context.Background(), "counter").Int64()
	if err != nil {
		log.Fatalf("Failed to get counter from Redis: %v", err)
	}

	// Find max
	var maxVal int64
	for _, v := range s.results {
		if v > maxVal {
			maxVal = v
		}
	}

	fmt.Printf("Max reported: %d, Redis counter: %d\n", maxVal, redisVal)
	// TODO: check also if equal with expected value
	if maxVal == redisVal {
		fmt.Println("SUCCESS: No race conditions detected.")
	} else {
		fmt.Println("FAILURE: Race condition detected!")
	}
}
