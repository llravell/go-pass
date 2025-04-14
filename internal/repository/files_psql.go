package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type FilesPostgresRepository struct {
	conn *sql.DB
}

func NewFilesPostgresRepository(conn *sql.DB) *FilesPostgresRepository {
	return &FilesPostgresRepository{
		conn: conn,
	}
}

func (repo *FilesPostgresRepository) UploadFile(
	ctx context.Context,
	userID int,
	file *entity.File,
	uploadFn func() (int64, error),
) error {
	return runInTx(ctx, repo.conn, func(tx *sql.Tx) error {
		var uploadStatus string

		row := tx.QueryRowContext(ctx, `
			SELECT upload_status
			FROM files
			WHERE user_id=$1 AND minio_bucket=$2 AND name=$3
			FOR UPDATE;
		`, userID, file.MinioBucket, file.Name)

		err := row.Scan(&uploadStatus)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if uploadStatus == "pending" {
			return entity.ErrFileAlreadyUploading
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO files (name, user_id, minio_bucket, upload_status)
				VALUES ($1, $2, $3, 'pending')
			ON CONFLICT (user_id, minio_bucket, name) DO UPDATE
			SET upload_status = EXCLUDED.upload_status;
		`, file.Name, userID, file.MinioBucket)
		if err != nil {
			return err
		}

		fileSize, err := uploadFn()
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE files
			SET upload_status='done', size=$1, meta=$2
			WHERE user_id=$3 AND minio_bucket=$4 AND name=$5
    `, fileSize, file.Meta, userID, file.MinioBucket, file.Name)
		if err != nil {
			return err
		}

		return nil
	})
}

func (repo *FilesPostgresRepository) GetFileByName(
	ctx context.Context,
	userID int,
	bucket string,
	name string,
) (*entity.File, error) {
	file := entity.File{Name: name, MinioBucket: bucket}

	row := repo.conn.QueryRowContext(ctx, `
		SELECT meta, size
		FROM files
		WHERE user_id=$1 AND minio_bucket=$2 AND name=$3 AND upload_status='done' AND NOT is_deleted;
	`, userID, bucket, name)

	err := row.Scan(&file.Meta, &file.Size)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
