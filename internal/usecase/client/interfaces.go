package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=../../mocks/mock_usecase_client.go -package=mocks

type (
	SessionRepository interface {
		GetSession(ctx context.Context) (*entity.ClientSession, error)
		SetSession(ctx context.Context, session *entity.ClientSession) error
	}
)
