package files

import (
	"context"
	"errors"
	"io"

	"rankode/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStorage struct {
	cl *minio.Client
}

type GetFileParams struct {
	Bucket string
	Name   string
}

type UploadFileParams struct {
	Bucket string
	Name   string
	File   io.Reader
	Size   int64
}

func NewFileStorage(cfg *config.Config) *FileStorage {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	return &FileStorage{cl: client}
}

func (s *FileStorage) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.cl.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = s.cl.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err == nil {
		return nil
	}

	exists, existsErr := s.cl.BucketExists(ctx, bucket)
	if existsErr == nil && exists {
		return nil
	}
	if existsErr != nil {
		return errors.Join(err, existsErr)
	}
	return err
}

func (s *FileStorage) UploadFile(ctx context.Context, params UploadFileParams) error {
	_, err := s.cl.PutObject(ctx, params.Bucket, params.Name, params.File, params.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

func (s *FileStorage) GetFile(ctx context.Context, params GetFileParams) (io.Reader, error) {
	file, err := s.cl.GetObject(ctx, params.Bucket, params.Name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return file, nil
}
