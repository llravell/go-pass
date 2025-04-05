package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type PasswordsPostgresRepository struct {
	conn *sql.DB
}

func NewPasswordsPostgresRepository(conn *sql.DB) *PasswordsPostgresRepository {
	return &PasswordsPostgresRepository{
		conn: conn,
	}
}

func (repo *PasswordsPostgresRepository) AddNewPassword(
	ctx context.Context,
	userID int,
	password *entity.Password,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		INSERT INTO passwords (name, encrypted_pass, meta, version, user_id)
		VALUES
			($1, $2, $3, $4, $5);
	`, password.Name, password.Value, password.Meta, password.Version, userID)

	if err != nil {
		return err
	}

	return nil
}

func (repo *PasswordsPostgresRepository) UpdateByName(
	ctx context.Context,
	userID int,
	name string,
	updateFn func(password *entity.Password) (*entity.Password, error),
) error {
	return runInTx(repo.conn, func(tx *sql.Tx) error {
		var pass entity.Password

		row := tx.QueryRowContext(ctx, `
			SELECT name, encrypted_pass, meta, version, is_deleted
			FROM passwords
			WHERE user_id=$1 name=$2
			FOR UPDATE;
		`, userID, name)

		err := row.Scan(&pass.Name, &pass.Value, &pass.Meta, &pass.Version, &pass.Deleted)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return entity.ErrPasswordDoesNotExist
			}

			return err
		}

		updatedPass, err := updateFn(&pass)
		if err != nil {
			return err
		}

		if updatedPass == nil {
			return nil
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE passwords
			SET encrypted_pass=$1, meta=$2, version=$3, is_deleted=$4
			WHERE user_id=$5 name=$6;
		`, pass.Value, pass.Meta, pass.Version, pass.Deleted, userID, pass.Name)
		if err != nil {
			return err
		}

		return nil
	})
}
