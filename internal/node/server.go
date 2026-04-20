// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/crolab/core/api/proto/v1"
)

// jobLogs stores log channels keyed by jobID so StreamLogs can read real output.
var jobLogs = struct {
	sync.RWMutex
	channels map[string]chan string
}{channels: make(map[string]chan string)}

func getOrCreateLogChan(jobID string) chan string {
	jobLogs.Lock()
	defer jobLogs.Unlock()
	if ch, ok := jobLogs.channels[jobID]; ok {
		return ch
	}
	ch := make(chan string, 256)
	jobLogs.channels[jobID] = ch
	return ch
}

func cleanupLogChan(jobID string) {
	jobLogs.Lock()
	defer jobLogs.Unlock()
	delete(jobLogs.channels, jobID)
}

// MaxConcurrentJobs controls how many Docker containers run simultaneously.
var MaxConcurrentJobs = 2

// QueueTimeout is how long a job waits in queue before being rejected.
var QueueTimeout = 5 * time.Minute

var jobSemaphore chan struct{}

func initSemaphore() {
	if jobSemaphore == nil {
		jobSemaphore = make(chan struct{}, MaxConcurrentJobs)
	}
}

type crolabServer struct {
	pb.UnimplementedCrolabServiceServer
}

// SubmitJob receives the serialized job, decodes it, and triggers DockerRunner.
func (s *crolabServer) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	initSemaphore()
	incTotal()

	log.Printf("📥 Job recebido — imagem: %s, comando: %s", req.Image, req.Command)

	jobID := uuid.New().String()
	logChan := getOrCreateLogChan(jobID)

	// Try to acquire a slot, or queue with timeout
	select {
	case jobSemaphore <- struct{}{}:
		// Got a slot immediately
	default:
		// All slots busy — queue the job
		incQueued()
		logChan <- fmt.Sprintf("Fila: aguardando slot (máx %d simultâneos)...\n", MaxConcurrentJobs)
		log.Printf("⏳ Job %s enfileirado (slots cheios)", jobID)

		select {
		case jobSemaphore <- struct{}{}:
			decQueued()
		case <-time.After(QueueTimeout):
			decQueued()
			close(logChan)
			cleanupLogChan(jobID)
			return nil, status.Errorf(codes.ResourceExhausted,
				"Timeout na fila (%v). Todos os %d slots ocupados.", QueueTimeout, MaxConcurrentJobs)
		}
	}

	go func() {
		incRunning()
		defer func() {
			<-jobSemaphore
			decRunning()
		}()
		defer close(logChan)
		defer cleanupLogChan(jobID)

		logChan <- fmt.Sprintf("Job %s iniciando...\n", jobID)

		err := RunDockerJob(jobID, req.Image, req.Command, req.Payload, logChan)
		if err != nil {
			incFailed()
			logChan <- fmt.Sprintf("ERRO: %v\n", err)
			log.Printf("❌ Job %s falhou: %v", jobID, err)
		} else {
			incCompleted()
			logChan <- "Job finalizado com sucesso.\n"
			log.Printf("✅ Job %s concluído", jobID)
		}
	}()

	return &pb.JobResponse{
		JobId:  jobID,
		Status: "QUEUED",
	}, nil
}

// StreamLogs reads real Docker output from the log channel for a given job.
func (s *crolabServer) StreamLogs(req *pb.LogRequest, stream pb.CrolabService_StreamLogsServer) error {
	logChan := getOrCreateLogChan(req.JobId)

	for msg := range logChan {
		if req.Abort {
			log.Printf("⛔ Abort solicitado para job %s", req.JobId)
			return nil
		}
		if err := stream.Send(&pb.LogMessage{Content: msg}); err != nil {
			return err
		}
	}
	return nil
}

// --- Auth Interceptors ---

func TokenAuthInterceptor(validToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Metadata ausente.")
		}
		values := md["authorization"]
		if len(values) == 0 || values[0] != validToken {
			return nil, status.Errorf(codes.Unauthenticated, "Token inválido.")
		}
		return handler(ctx, req)
	}
}

func StreamTokenAuthInterceptor(validToken string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "Metadata ausente.")
		}
		values := md["authorization"]
		if len(values) == 0 || values[0] != validToken {
			return status.Errorf(codes.Unauthenticated, "Token inválido.")
		}
		return handler(srv, ss)
	}
}

// Start launches the gRPC server with optional metrics HTTP endpoint.
func Start(port string, token string, tlsCert string, tlsKey string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("falha ao ouvir porta %s: %v", port, err)
	}

	// Start metrics HTTP server on port+1 offset
	StartMetricsServer(":9090")

	opts := []grpc.ServerOption{}

	if tlsCert != "" && tlsKey != "" {
		creds, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("falha ao carregar certificados TLS: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Printf("🔒 gRPC TLS Ativado")
	}

	if token != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(TokenAuthInterceptor(token)),
			grpc.StreamInterceptor(StreamTokenAuthInterceptor(token)),
		)
		log.Printf("🛡️  Crolab Node em %s (auth ativada, %d slots)", port, MaxConcurrentJobs)
	} else {
		log.Printf("⚠️  Crolab Node em %s (SEM auth, %d slots)", port, MaxConcurrentJobs)
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCrolabServiceServer(grpcServer, &crolabServer{})

	return grpcServer.Serve(lis)
}
