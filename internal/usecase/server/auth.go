package server

import (
	"context"
	"time"

	"github.com/llravell/go-pass/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo      UserRepository
	jwtIssuer JWTIssuer
}

func NewAuthUseCase(repo UserRepository, jwtIssuer JWTIssuer) *AuthUseCase {
	return &AuthUseCase{
		repo:      repo,
		jwtIssuer: jwtIssuer,
	}
}

func (auth *AuthUseCase) RegisterUser(ctx context.Context, login string, password string) (*entity.User, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return auth.repo.StoreUser(ctx, login, string(passwordBytes))
}

func (auth *AuthUseCase) BuildUserToken(user *entity.User, ttl time.Duration) (string, error) {
	return auth.jwtIssuer.Issue(user.ID, ttl)
}

func (auth *AuthUseCase) VerifyUser(ctx context.Context, login string, password string) (*entity.User, error) {
	user, err := auth.repo.FindUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}
