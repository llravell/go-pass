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

	JWTIssuer interface {
		Issue(userID int, ttl time.Duration) (string, error)
	}
)
