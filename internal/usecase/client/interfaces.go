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
	PasswordsRepository interface {
		PasswordExists(ctx context.Context, name string) (bool, error)
		GetPasswordByName(ctx context.Context, name string) (*entity.Password, error)
		CreateNewPassword(ctx context.Context, password *entity.Password) error
		UpdatePassword(ctx context.Context, password *entity.Password) error
		GetPasswords(ctx context.Context) ([]*entity.Password, error)
		DeletePasswordHard(ctx context.Context, name string) error
		DeletePasswordSoft(ctx context.Context, name string) error
	}
)
