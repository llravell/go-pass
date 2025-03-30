package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	sessionRepo SessionRepository
	authClient  pb.AuthClient
}

func NewAuthUseCase(
	sessionRepo SessionRepository,
	authClient pb.AuthClient,
) *AuthUseCase {
	return &AuthUseCase{
		sessionRepo: sessionRepo,
		authClient:  authClient,
	}
}

func (auth *AuthUseCase) Register(
	ctx context.Context,
	login, password string,
) error {
	resp, err := auth.authClient.Register(ctx, &pb.AuthRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	return auth.saveUserSession(ctx, login, password, resp.Token)
}

func (auth *AuthUseCase) Login(
	ctx context.Context,
	login, password string,
) error {
	resp, err := auth.authClient.Login(ctx, &pb.AuthRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	return auth.saveUserSession(ctx, login, password, resp.Token)
}

func (auth *AuthUseCase) ValidateMasterPassword(
	ctx context.Context,
	masterPassword string,
) error {
	session, err := auth.sessionRepo.GetSession(ctx)
	if err != nil {
		return err
	}

	if len(session.MasterPassHash) == 0 {
		return entity.ErrNoSession
	}

	return bcrypt.CompareHashAndPassword([]byte(session.MasterPassHash), []byte(masterPassword))
}

func (auth *AuthUseCase) saveUserSession(
	ctx context.Context,
	login, password, authToken string,
) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return auth.sessionRepo.SetSession(ctx, &entity.ClientSession{
		Login:          login,
		MasterPassHash: string(passHash),
		AuthToken:      authToken,
	})
}
