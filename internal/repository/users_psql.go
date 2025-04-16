package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/llravell/go-pass/internal/entity"
)

type UsersRepository struct {
	conn *sql.DB
}

func NewUsersRepository(conn *sql.DB) *UsersRepository {
	return &UsersRepository{conn: conn}
}

func (r *UsersRepository) StoreUser(
	ctx context.Context,
	login string,
	passwordHash string,
) (*entity.User, error) {
	var user entity.User

	row := r.conn.QueryRowContext(ctx, `
		INSERT INTO users (login, password)
		VALUES
			($1, $2)
		RETURNING id, login, password;
	`, login, passwordHash)

	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = entity.ErrUserConflict
		}
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UsersRepository) FindUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	var user entity.User

	row := r.conn.QueryRowContext(ctx, `
		SELECT id, login, password
		FROM users
		WHERE
			login=$1;
	`, login)

	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
