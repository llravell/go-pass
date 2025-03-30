package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type PasswordsSqliteRepository struct {
	conn *sql.DB
}

func NewPasswordsSqliteRepository(conn *sql.DB) *PasswordsSqliteRepository {
	return &PasswordsSqliteRepository{
		conn: conn,
	}
}

func (repo *PasswordsSqliteRepository) GetPasswords(
	ctx context.Context,
) ([]*entity.Password, error) {
	var passwords []*entity.Password

	rows, err := repo.conn.QueryContext(ctx, `
		SELECT name, encrypted_pass, meta, version, is_deleted
		FROM passwords;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var pass entity.Password

		err = rows.Scan(&pass.Name, &pass.Value, &pass.Meta, &pass.Version, &pass.Deleted)
		if err != nil {
			return nil, err
		}

		passwords = append(passwords, &pass)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return passwords, nil
		}

		return nil, err
	}

	return passwords, nil
}

func (repo *PasswordsSqliteRepository) GetPasswordByName(
	ctx context.Context,
	name string,
) (*entity.Password, error) {
	var pass entity.Password

	row := repo.conn.QueryRowContext(ctx, `
		SELECT name, encrypted_pass, meta, version, is_deleted
		FROM passwords
		WHERE name=? AND NOT is_deleted;
	`, name)

	err := row.Scan(&pass.Name, &pass.Value, &pass.Meta, &pass.Version, &pass.Deleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrPasswordDoesNotExist
		}

		return nil, err
	}

	return &pass, nil
}

func (repo *PasswordsSqliteRepository) PasswordExists(
	ctx context.Context,
	name string,
) (bool, error) {
	var passwordId int

	row := repo.conn.QueryRowContext(ctx, `
		SELECT id
		FROM passwords
		WHERE name=? AND NOT is_deleted;
	`, name)

	err := row.Scan(&passwordId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return passwordId > 0, nil
}

func (repo *PasswordsSqliteRepository) CreateNewPassword(
	ctx context.Context,
	password *entity.Password,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		INSERT INTO passwords (name, encrypted_pass, meta, version)
		VALUES
			(?, ?, ?, ?);
	`, password.Name, password.Value, password.Meta, password.Version)

	if err != nil {
		return err
	}

	return nil
}

func (repo *PasswordsSqliteRepository) UpdatePassword(
	ctx context.Context,
	password *entity.Password,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		UPDATE passwords
		SET encrypted_pass=?, meta=?, version=?
		WHERE name=?;
	`, password.Value, password.Meta, password.Version, password.Name)

	if err != nil {
		return err
	}

	return nil
}
