package security_test

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/crolab/core/api/proto/v1"
	"github.com/crolab/core/internal/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func startSecureNode(t *testing.T, port string, token string) {
	node.MaxConcurrentJobs = 2
	go func() {
		// Mock TLS as passing nil uses only insecure fallback if testing in process
		_ = node.Start(port, token, "", "")
	}()
	time.Sleep(100 * time.Millisecond)
}

func getFreePort() string {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := l.Addr().String()
	l.Close()
	return port
}

func TestAuthBypassPrevention(t *testing.T) {
	port := getFreePort()
	secretToken := "very-secret-token"

	startSecureNode(t, port, secretToken)

	conn, err := grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Erro inesperado ao conectar local: %v", err)
	}
	defer conn.Close()

	client := pb.NewCrolabServiceClient(conn)

	// Tentativa 1: Sem Token (Auth Bypass de metadados ausente)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.SubmitJob(ctx, &pb.JobRequest{Image: "ubuntu", Command: "ls"})
	if err == nil {
		t.Fatalf("Vulnerabilidade Crítica! Conexão sem token aceita incondicionalmente no gRPC Payload.")
	}
	if !strings.Contains(err.Error(), "Metadata ausente") && !strings.Contains(err.Error(), "Unauthenticated") {
		t.Errorf("Esperado Unauthenticated ou Metadata ausente, recebido: %v", err)
	}

	// Tentativa 2: Token Incorreto ou Truncado (Fuzzing simples)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	ctx2 = metadata.AppendToOutgoingContext(ctx2, "authorization", "wrong-token-hacker")

	_, err = client.SubmitJob(ctx2, &pb.JobRequest{Image: "ubuntu", Command: "ls"})
	if err == nil {
		t.Fatalf("Vulnerabilidade Crítica! Conexão com Token Errado injetada sob Payload malicioso!")
	}
	if !strings.Contains(err.Error(), "Token inválido") && !strings.Contains(err.Error(), "Unauthenticated") {
		t.Errorf("Esperado Token inválido, recebido: %v", err)
	}

	// Tentativa 3: Login Sucesso
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel3()
	ctx3 = metadata.AppendToOutgoingContext(ctx3, "authorization", secretToken)

	_, err = client.SubmitJob(ctx3, &pb.JobRequest{Image: "python:3.11-slim", Command: "echo ok", Payload: []byte("PK\x03\x04")})
	// A rota SubmitJob retornará "caminho suspeito" porque enviamos zip truncado (zipslip barrou a payload) ou process erro.
	// O importante é que a Auth passou!
	if err != nil && strings.Contains(err.Error(), "Unauthenticated") {
		t.Errorf("Login legítimo bloqueado injustamente: %v", err)
	}
}
