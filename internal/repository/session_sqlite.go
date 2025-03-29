package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

const (
	loginKey      = "login"
	masterPassKey = "master_password"
	authTokenKey  = "auth_token"
)

type SessionSqliteRepository struct {
	conn *sql.DB
}

func NewSessionSqliteRepository(conn *sql.DB) *SessionSqliteRepository {
	return &SessionSqliteRepository{
		conn: conn,
	}
}

func (repo *SessionSqliteRepository) GetSession(
	ctx context.Context,
) (*entity.ClientSession, error) {
	var session entity.ClientSession

	loginRow := repo.conn.QueryRowContext(ctx, "SELECT value FROM session WHERE key=?", loginKey)
	err := loginRow.Scan(&session.Login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	passRow := repo.conn.QueryRowContext(ctx, "SELECT value FROM session WHERE key=?", masterPassKey)
	err = passRow.Scan(&session.MasterPassHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &session, nil
}

func (repo *SessionSqliteRepository) SetSession(
	ctx context.Context,
	session *entity.ClientSession,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		INSERT OR REPLACE INTO session (key, value)
		VALUES
			(?, ?),
			(?, ?);
			(?, ?);
	`,
		loginKey, session.Login,
		masterPassKey, session.MasterPassHash,
		authTokenKey, session.AuthToken,
	)

	return err
}
