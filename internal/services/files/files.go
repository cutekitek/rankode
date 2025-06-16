package files

import (
	"context"
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
	Name string
}

type UploadFileParams struct {
	Bucket string
	Name string
	File io.Reader
	Size int64
}

func NewFileStorage(cfg *config.Config) *FileStorage {
	client, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds: credentials.NewStaticV4(cfg.MinIOLogin, cfg.MinIOPassword, ""),
	})
	if err != nil{
		panic(err)
	}
	return &FileStorage{cl: client} 
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