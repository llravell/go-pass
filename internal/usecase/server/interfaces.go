package server

import (
	"context"
	"time"

	"github.com/llravell/go-pass/internal/entity"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=../../mocks/mock_usecase_server.go -package=mocks

type (
	UserRepository interface {
		StoreUser(ctx context.Context, login string, password string) (*entity.User, error)
		FindUserByLogin(ctx context.Context, login string) (*entity.User, error)
	}

	PasswordsRepository interface {
		UpdateByName(
			ctx context.Context,
			userID int,
			name string,
			updateFn func(password *entity.Password) (*entity.Password, error),
		) error
		AddNewPassword(ctx context.Context, userID int, password *entity.Password) error
		DeletePasswordByName(ctx context.Context, userID int, name string) error
		GetPasswords(ctx context.Context, userID int) ([]*entity.Password, error)
	}

	JWTIssuer interface {
		Issue(userID int, ttl time.Duration) (string, error)
	}
)
