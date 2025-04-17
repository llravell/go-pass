package server

import (
	"context"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/rs/zerolog"
)

type FilesUseCase struct {
	repo           FilesRepository
	s3Storage      FilesS3Storage
	fileDeletingWP FileDeletingWorkerPool
	log            *zerolog.Logger
}

func NewFilesUseCase(
	repo FilesRepository,
	s3Storage FilesS3Storage,
	fileDeletingWP FileDeletingWorkerPool,
	log *zerolog.Logger,
) *FilesUseCase {
	return &FilesUseCase{
		repo:           repo,
		s3Storage:      s3Storage,
		fileDeletingWP: fileDeletingWP,
		log:            log,
	}
}

func (uc *FilesUseCase) UploadFile(
	ctx context.Context,
	userID int,
	file *entity.File,
	fileReader io.Reader,
) error {
	return uc.repo.UploadFile(ctx, userID, file, func() (int64, error) {
		return uc.s3Storage.UploadFile(ctx, file, fileReader)
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

	return uc.s3Storage.DownloadFile(ctx, file)
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

	err = uc.fileDeletingWP.QueueWork(&FileDeleteWork{
		file:      &entity.File{Name: name, MinioBucket: bucket},
		s3Storage: uc.s3Storage,
		log:       uc.log,
	})
	if err != nil {
		return err
	}

	return nil
}
