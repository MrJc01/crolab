package pool_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	pb "github.com/crolab/core/api/proto/v1"
	"google.golang.org/grpc"
)

type mockGrpcServer struct {
	pb.UnimplementedCrolabServiceServer
}

func (s *mockGrpcServer) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	return &pb.JobResponse{JobId: "mock-job-123", Status: "RUNNING"}, nil
}

func (s *mockGrpcServer) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	stream.Send(&pb.LogMessage{Content: "Log failover ok\n"})
	return nil
}

func startMockNode(t *testing.T, port string) (*grpc.Server, string) {
	t.Helper()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterCrolabServiceServer(srv, &mockGrpcServer{})
	go srv.Serve(lis)
	return srv, lis.Addr().String()
}

func TestPoolFailoverCascade(t *testing.T) {
	node3, node3Addr := startMockNode(t, "127.0.0.1:0")
	defer node3.Stop()
	time.Sleep(50 * time.Millisecond)

	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Port bind failed: %v", err)
	}
	defer lis.Close()
	dynamicPort := lis.Addr().String()

	os.MkdirAll(homeDir+"/.crolab", 0755)
	os.WriteFile(homeDir+"/.crolab/config.yaml", []byte(`
cloud_token: "mock-token-123"
cloud_api: "http://`+dynamicPort+`"
`), 0644)

	env := append(os.Environ(), "HOME="+homeDir, "CROLAB_HOME="+homeDir)

	mockCloud := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/client/run" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"job_id":  "job123",
					"message": "Ticket gerado",
					"nodes": []map[string]string{
						{"address": "127.0.0.1:49991", "token": "t1"},
						{"address": "127.0.0.1:49992", "token": "t2"},
						{"address": node3Addr, "token": "t3"},
					},
					"status": "running",
				})
			}
		}),
	}
	go mockCloud.Serve(lis)
	defer mockCloud.Close()
	time.Sleep(50 * time.Millisecond)

	binDir := t.TempDir()
	binPath := binDir + "/crolab.bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/crolab/")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v\n%s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "run", tmpDir, "--plan", "start")
	cmd.Env = env
	
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("Comando Crolab CLI travou e atingiu o timeout do teste! Output:\n%s", string(out))
	}
	
	outputStr := string(out)
	
	if !strings.Contains(outputStr, "Ligando Node 1") {
		t.Errorf("Não tentou ligar no node 1\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Ligando Node 2") {
		t.Errorf("Não tentou ligar no node 2\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Ligando Node ") {
		t.Errorf("Não tentou ligar no node 3\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Log failover ok") {
		t.Errorf("Não registrou log stream do node 3 fallback\n%s", outputStr)
	}
}
