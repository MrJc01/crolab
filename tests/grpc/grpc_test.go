package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/crolab/core/internal/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/crolab/core/api/proto/v1"
)

type mockSrv struct {
	pb.UnimplementedCrolabServiceServer
	lastImage   string
	lastCommand string
}

func (s *mockSrv) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	s.lastImage = req.Image
	s.lastCommand = req.Command
	return &pb.JobResponse{JobId: "test-123", Status: "QUEUED"}, nil
}

func (s *mockSrv) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	stream.Send(&pb.LogMessage{Content: "log output\n"})
	return nil
}

func startMockServer(t *testing.T, port string) (*mockSrv, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &mockSrv{}
	grpcServer := grpc.NewServer()
	pb.RegisterCrolabServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)
	time.Sleep(50 * time.Millisecond)
	return srv, func() { grpcServer.Stop() }
}

func TestSubmitJobSendsPayload(t *testing.T) {
	_, cleanup := startMockServer(t, ":15001")
	defer cleanup()

	tmp := t.TempDir()
	err := cli.SubmitJob("127.0.0.1:15001", "", "python:3.11", "python train.py", tmp, false)
	if err != nil {
		t.Fatalf("SubmitJob falhou: %v", err)
	}
}

func TestSubmitJobImageAndCommand(t *testing.T) {
	srv, cleanup := startMockServer(t, ":15002")
	defer cleanup()

	tmp := t.TempDir()
	cli.SubmitJob("127.0.0.1:15002", "", "nvidia/cuda:12", "python main.py", tmp, false)

	if srv.lastImage != "nvidia/cuda:12" {
		t.Errorf("imagem errada: %s", srv.lastImage)
	}
	if srv.lastCommand != "python main.py" {
		t.Errorf("comando errado: %s", srv.lastCommand)
	}
}

func TestSubmitJobBadAddress(t *testing.T) {
	tmp := t.TempDir()
	err := cli.SubmitJob("127.0.0.1:19999", "", "alpine", "echo", tmp, false)
	if err == nil {
		t.Error("deveria falhar com endereço inexistente")
	}
}

func TestGRPCMetadataAuth(t *testing.T) {
	lis, err := net.Listen("tcp", ":15003")
	if err != nil {
		t.Fatal(err)
	}

	var capturedToken string
	authServer := &mockSrv{}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				vals := md["authorization"]
				if len(vals) > 0 {
					capturedToken = vals[0]
				}
			}
			return handler(ctx, req)
		}),
	)
	pb.RegisterCrolabServiceServer(grpcServer, authServer)
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()
	time.Sleep(50 * time.Millisecond)

	tmp := t.TempDir()
	cli.SubmitJob("127.0.0.1:15003", "meu-token-secreto", "alpine", "echo", tmp, false)

	if capturedToken != "meu-token-secreto" {
		t.Errorf("token não chegou no server: '%s'", capturedToken)
	}
}

func TestGRPCConnectionTimeout(t *testing.T) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "10.255.255.1:4422",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	elapsed := time.Since(start)

	if err == nil {
		conn.Close()
		t.Error("deveria ter timeout")
	}
	if elapsed > 2*time.Second {
		t.Errorf("timeout demorou demais: %v", elapsed)
	}
}
