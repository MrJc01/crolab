// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/crolab/core/api/proto/v1"
)

// SubmitJob packs the current dir and sends to grpc node
func SubmitJob(serverAddr, token, image, command, targetDir string, useTls bool) error {
	payload, err := ZipDir(targetDir)
	if err != nil {
		return fmt.Errorf("falha ao comprimir workspace %s: %v", targetDir, err)
	}

	creds := insecure.NewCredentials()
	if useTls {
		creds = credentials.NewClientTLSFromCert(nil, "")
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, serverAddr, 
		grpc.WithTransportCredentials(creds), 
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("node offline ou restrito: %v", err)
	}
	defer conn.Close()

	client := pb.NewCrolabServiceClient(conn)

	log.Printf("📦 Roteando [%d bytes] -> %s", len(payload), serverAddr)
	req := &pb.JobRequest{
		Image:   image,
		Command: command,
		Payload: payload,
	}

	// Adentra Auth Metadados e Timeout Rápido
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}

	resp, err := client.SubmitJob(ctx, req)
	if err != nil {
		return fmt.Errorf("🚫 Acesso Rejeitado pela Nuvem (Auth?): %v", err)
	}

	log.Printf("✅ Job Escalonado: %s", resp.JobId)
	return tailLogs(client, resp.JobId, token)
}

func tailLogs(client pb.CrolabServiceClient, jobID string, token string) error {
	ctx := context.Background()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}

	stream, err := client.StreamLogs(ctx, &pb.LogRequest{JobId: jobID})
	if err != nil {
		return fmt.Errorf("failed to tail logs: %v", err)
	}

	fmt.Println("--- REMOTE LOGS ---")
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream interrupted: %v", err)
		}
		
		fmt.Print(msg.Content)
	}
	fmt.Println("--- JOB FINISHED ---")
	return nil
}

func ZipDir(src string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, f)
		return err
	})

	if err != nil {
		return nil, err
	}
	zipWriter.Close()
	return buf.Bytes(), nil
}
