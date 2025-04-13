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
