package server

import (
	"context"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/minio/minio-go/v7"
)

type FilesUseCase struct {
	repo        FilesRepository
	minioClient *minio.Client
}

func NewFilesUseCase(repo FilesRepository, minioClient *minio.Client) *FilesUseCase {
	return &FilesUseCase{
		repo:        repo,
		minioClient: minioClient,
	}
}

func (uc *FilesUseCase) UploadFile(
	ctx context.Context,
	userID int,
	file *entity.File,
	fileReader io.Reader,
) error {
	return uc.repo.UploadFile(ctx, userID, file, func() (int64, error) {
		info, err := uc.minioClient.PutObject(
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
	})
}

func (uc *FilesUseCase) DownloadFile(
	ctx context.Context,
	userID int,
	bucket string,
	name string,
) (io.ReadCloser, error) {
	file, err := uc.repo.GetFileByName(ctx, userID, bucket, name)
	if err != nil {
		return nil, err
	}

	obj, err := uc.minioClient.GetObject(ctx, file.MinioBucket, file.Name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (uc *FilesUseCase) GetFiles(
	ctx context.Context,
	userID int,
	bucket string,
) ([]*entity.File, error) {
	return uc.repo.GetFiles(ctx, userID, bucket)
}

func (uc *FilesUseCase) DeleteFile(
	ctx context.Context,
	userID int,
	bucket string,
	name string,
) error {
	err := uc.repo.DeleteFileByName(ctx, userID, bucket, name)
	if err != nil {
		return err
	}

	return nil
}
