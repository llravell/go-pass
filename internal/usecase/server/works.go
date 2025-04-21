package server

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/rs/zerolog"
)

type FileDeleteWork struct {
	log       *zerolog.Logger
	file      *entity.File
	s3Storage FilesS3Storage
}

func (w *FileDeleteWork) Do(ctx context.Context) {
	err := w.s3Storage.DeleteFile(ctx, w.file)
	if err != nil {
		w.log.Error().Err(err).Msg("minio file deleting failed")
	}
}
