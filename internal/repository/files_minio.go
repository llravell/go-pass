package repository

import (
	"context"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/minio/minio-go/v7"
)

type FilesMinioStorage struct {
	client *minio.Client
}

func NewFilesMinioStorage(client *minio.Client) *FilesMinioStorage {
	return &FilesMinioStorage{
		client: client,
	}
}

func (repo *FilesMinioStorage) UploadFile(
	ctx context.Context,
	file *entity.File,
	fileReader io.Reader,
) (int64, error) {
	info, err := repo.client.PutObject(
		ctx,
		file.MinioBucket,
		file.Name,
		fileReader,
		int64(-1),
		minio.PutObjectOptions{},
	)
	if err != nil {
		return 0, err
	}

	return info.Size, nil
}

func (repo *FilesMinioStorage) DownloadFile(
	ctx context.Context,
	file *entity.File,
) (io.ReadCloser, error) {
	obj, err := repo.client.GetObject(ctx, file.MinioBucket, file.Name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (repo *FilesMinioStorage) DeleteFile(
	ctx context.Context,
	file *entity.File,
) error {
	return repo.client.RemoveObject(
		ctx,
		file.MinioBucket,
		file.Name,
		minio.RemoveObjectOptions{},
	)
}
