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
	CardsRepository interface {
		UpdateByName(
			ctx context.Context,
			userID int,
			name string,
			updateFn func(password *entity.Card) (*entity.Card, error),
		) error
		AddNewCard(ctx context.Context, userID int, card *entity.Card) error
		DeleteCardByName(ctx context.Context, userID int, name string) error
		GetCards(ctx context.Context, userID int) ([]*entity.Card, error)
	}
	FilesRepository interface {
		UploadFile(
			ctx context.Context,
			userID int,
			file *entity.File,
			uploadFn func() (int64, error),
		) error
		GetFileByName(
			ctx context.Context,
			userID int,
			bucket string,
			name string,
		) (*entity.File, error)
		GetFiles(
			ctx context.Context,
			userID int,
			bucket string,
		) ([]*entity.File, error)
		DeleteFileByName(
			ctx context.Context,
			userID int,
			bucket string,
			name string,
		) error
	}
	JWTIssuer interface {
		Issue(userID int, ttl time.Duration) (string, error)
	}
)
