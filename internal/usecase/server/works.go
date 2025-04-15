package server

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
)

type FileDeleteWork struct {
	log         *zerolog.Logger
	file        *entity.File
	minioClient *minio.Client
}

func (w *FileDeleteWork) Do(ctx context.Context) {
	err := w.minioClient.RemoveObject(
		ctx,
		w.file.MinioBucket,
		w.file.Name,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		w.log.Error().Err(err).Msg("minio file deleting failed")
	}
}
