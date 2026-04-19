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

// crolabServer mock for load tests to bypass Docker daemon dependency while testing P2P sockets
type mockServer struct {
	pb.UnimplementedCrolabServiceServer
}

func (s *mockServer) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	return &pb.JobResponse{JobId: "chaos", Status: "QUEUED"}, nil
}

func (s *mockServer) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	stream.Send(&pb.LogMessage{Content: "Chaos Test Log Output\n"})
	return nil
}

func TestChaosLoopSRE(t *testing.T) {
	// Spin up a mock server for extreme load testing on the RPC sockets without frying the host Docker Daemon.
	// If we used the real Docker runner in 50 tight loops, we would hit OS limitations instantly.
	lis, err := net.Listen("tcp", ":5599")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCrolabServiceServer(grpcServer, &mockServer{})

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	// Give it a millisecond to bind
	time.Sleep(100 * time.Millisecond)

	cycles := 50
	successCount := 0

	tmpDir := t.TempDir()

	for i := 0; i < cycles; i++ {
		t.Logf("SRE Stress Cycle -> [%d/%d]", i+1, cycles)
		
		err := cli.SubmitJob("127.0.0.1:5599", "DUMMY_TOKEN", "alpine:latest", "echo 'Chaos'", tmpDir)
		if err != nil {
			t.Errorf("Ciclo %d apresentou latência ou falha gRPC: %v", i, err)
		} else {
			successCount++
		}
	}

	if successCount != cycles {
		t.Fatalf("Esperava %d pacotes de sucessos sem lock, recebi %d. O RPC gargalou.", cycles, successCount)
	} else {
		t.Logf("Todos os %d ciclos de estresse completaram em tempo O(1) com sucesso no socket.", cycles)
	}
}
