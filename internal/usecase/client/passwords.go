package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
)

type PasswordsUseCase struct {
	passwordsRepo   PasswordsRepository
	passwordsClient pb.PasswordsClient
}

func NewPasswordsUseCase(
	passwordsRepo PasswordsRepository,
) *PasswordsUseCase {
	return &PasswordsUseCase{
		passwordsRepo: passwordsRepo,
	}
}

func (p *PasswordsUseCase) AddNewPassword(
	ctx context.Context,
	password entity.Password,
) error {
	exists, err := p.passwordsRepo.PasswordExists(ctx, password.Name)
	if err != nil {
		return err
	}

	if exists {
		return entity.ErrPasswordAlreadyExist
	}

	password.Version = 1

	response, err := p.passwordsClient.Sync(ctx, password.ToPB())
	if err != nil {
		return p.passwordsRepo.CreateNewPassword(ctx, &password)
	}

	if response.GetSuccess() {
		return p.passwordsRepo.CreateNewPassword(ctx, &password)
	}

	return entity.NewPasswordConflictErrorFromPB(response.GetConflict())
}

func (p *PasswordsUseCase) GetPasswordByName(
	ctx context.Context,
	name string,
) (*entity.Password, error) {
	return p.passwordsRepo.GetPasswordByName(ctx, name)
}

func (p *PasswordsUseCase) UpdatePassword(
	ctx context.Context,
	password *entity.Password,
) error {
	response, err := p.passwordsClient.Sync(ctx, password.ToPB())
	if err != nil {
		return p.passwordsRepo.UpdatePassword(ctx, password)
	}

	if response.GetSuccess() {
		return p.passwordsRepo.UpdatePassword(ctx, password)
	}

	return entity.NewPasswordConflictErrorFromPB(response.GetConflict())
}

func (p *PasswordsUseCase) GetList(
	ctx context.Context,
) ([]*entity.Password, error) {
	return p.passwordsRepo.GetPasswords(ctx)
}
