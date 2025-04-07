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

func (repo *PasswordsPostgresRepository) GetPasswords(
	ctx context.Context,
	userID int,
) ([]*entity.Password, error) {
	passwords := make([]*entity.Password, 0)

	rows, err := repo.conn.QueryContext(ctx, `
		SELECT name, encrypted_pass, meta, version
		FROM passwords
		WHERE user_id=$1 AND NOT is_deleted;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var password entity.Password

		err = rows.Scan(&password.Name, &password.Value, &password.Meta, &password.Version)
		if err != nil {
			return nil, err
		}

		passwords = append(passwords, &password)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return passwords, nil
		}

		return nil, err
	}

	return passwords, nil
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

func (repo *PasswordsPostgresRepository) DeletePasswordByName(
	ctx context.Context,
	userID int,
	name string,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		UPDATE passwords
		SET is_deleted=TRUE
		WHERE user_id=$1 AND name=$2;
	`, userID, name)
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
			WHERE user_id=$1 AND name=$2
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
			WHERE user_id=$5 AND name=$6;
		`, updatedPass.Value, updatedPass.Meta, updatedPass.Version, updatedPass.Deleted, userID, pass.Name)
		if err != nil {
			return err
		}

		return nil
	})
}
