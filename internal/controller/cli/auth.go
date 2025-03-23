package cli

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type SessionRepo interface {
	GetSession(ctx context.Context) (*entity.ClientSession, error)
	SetSession(ctx context.Context, session *entity.ClientSession) error
}

type AuthController struct {
	sessionRepo SessionRepo
}

func NewAuthController(sessionRepo SessionRepo) *AuthController {
	return &AuthController{
		sessionRepo: sessionRepo,
	}
}

func (c *AuthController) Register(ctx context.Context, login, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = c.sessionRepo.SetSession(ctx, &entity.ClientSession{
		Login:          login,
		MasterPassHash: string(hash),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *AuthController) Login(ctx context.Context, login, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = c.sessionRepo.SetSession(ctx, &entity.ClientSession{
		Login:          login,
		MasterPassHash: string(hash),
	})

	if err != nil {
		return err
	}

	return nil
}
