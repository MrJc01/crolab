package load

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/crolab/core/internal/cli"
	"google.golang.org/grpc"

	pb "github.com/crolab/core/api/proto/v1"
)

// Mock server for load tests — no Docker dependency.
type mockServer struct {
	pb.UnimplementedCrolabServiceServer
}

func (s *mockServer) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	return &pb.JobResponse{JobId: "load-test-job", Status: "QUEUED"}, nil
}

func (s *mockServer) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	stream.Send(&pb.LogMessage{Content: "mock log line\n"})
	return nil
}

func TestChaosLoopSRE(t *testing.T) {
	lis, err := net.Listen("tcp", ":5599")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCrolabServiceServer(grpcServer, &mockServer{})

	go func() { grpcServer.Serve(lis) }()
	defer grpcServer.Stop()

	time.Sleep(100 * time.Millisecond)

	cycles := 50
	successCount := 0
	tmpDir := t.TempDir()

	for i := 0; i < cycles; i++ {
		err := cli.SubmitJob("127.0.0.1:5599", "", "alpine:latest", "echo ok", tmpDir, false)
		if err != nil {
			t.Errorf("ciclo %d falhou: %v", i+1, err)
		} else {
			successCount++
		}
	}

	if successCount != cycles {
		t.Fatalf("esperava %d sucessos, obteve %d", cycles, successCount)
	}
	t.Logf("✓ %d ciclos completaram com sucesso (0 falhas)", cycles)
}
