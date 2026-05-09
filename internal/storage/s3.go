package storage

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// CloudStorage fornece a interface de persistência de Notebooks em S3/MinIO
type CloudStorage struct {
	client     *minio.Client
	bucketName string
}

// NewCloudStorage inicializa a conexão com o S3
func NewCloudStorage(endpoint, accessKey, secretKey, bucket string) (*CloudStorage, error) {
	// Se estiver usando AWS S3 verdadeiro, useSSL=true. Para MinIO local useSSL=false.
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &CloudStorage{
		client:     minioClient,
		bucketName: bucket,
	}, nil
}

// EnsureBucket verifica se o bucket do usuário/workspace existe. Cria se não.
func (s *CloudStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("Bucket %s criado com sucesso.", s.bucketName)
	}
	return nil
}

// SaveNotebook persiste um objeto JSON (.ipynb) na nuvem
func (s *CloudStorage) SaveNotebook(ctx context.Context, objectName string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	return err
}

// LoadNotebook carrega um notebook salvo
func (s *CloudStorage) LoadNotebook(ctx context.Context, objectName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}
