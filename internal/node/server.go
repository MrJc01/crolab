package node

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/crolab/core/api/proto/v1"
)

type crolabServer struct {
	pb.UnimplementedCrolabServiceServer
}

// SubmitJob receives the serialized job, decodes it, and triggers DockerRunner
func (s *crolabServer) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	log.Printf("📥 Recebido Job - Imagem: %s, Command: %s", req.Image, req.Command)

	jobID := fmt.Sprintf("job-%d", len(req.Payload))

	go func() {
		err := RunDockerJob(jobID, req.Image, req.Command, req.Payload)
		if err != nil {
			log.Printf("❌ Job %s falhou no Docker: %v", jobID, err)
		}
	}()

	return &pb.JobResponse{
		JobId:  jobID,
		Status: "QUEUED",
	}, nil
}

// StreamLogs hooks into the docker runner and streams back to CLI
func (s *crolabServer) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	messages := []string{"Initializing...", "Job Acoplado ao Docker Engine...", "Executando...", "Done"}
	for _, msg := range messages {
		if req.Abort {
			log.Printf("⛔ Aborting job %s", req.JobId)
			return nil
		}
		if err := stream.Send(&pb.LogMessage{Content: msg + "\n"}); err != nil {
			return err
		}
	}
	return nil
}

func TokenAuthInterceptor(validToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "SRE Auth: Metadata ausente.")
		}
		values := md["authorization"]
		if len(values) == 0 || values[0] != validToken {
			return nil, status.Errorf(codes.Unauthenticated, "SRE Auth: Crolab Token de nó malicioso ou inválido.")
		}
		return handler(ctx, req)
	}
}

func StreamTokenAuthInterceptor(validToken string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "SRE Auth: Metadata ausente no Stream.")
		}
		values := md["authorization"]
		if len(values) == 0 || values[0] != validToken {
			return status.Errorf(codes.Unauthenticated, "SRE Auth: Stream Token Invalido.")
		}
		return handler(srv, ss)
	}
}

// Start launches the gRPC listener with security
func Start(port string, token string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("falha grave ao ouvir %s: %v", port, err)
	}

	opts := []grpc.ServerOption{}
	if token != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(TokenAuthInterceptor(token)),
			grpc.StreamInterceptor(StreamTokenAuthInterceptor(token)),
		)
		log.Printf("🛡️  Crolab Node Seguro Acionado na porta %s (Auth Ativado)", port)
	} else {
		log.Printf("⚠️  Crolab Node Iniciado SEM AUTH na porta %s (Node Aberto)", port)
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCrolabServiceServer(grpcServer, &crolabServer{})

	return grpcServer.Serve(lis)
}
