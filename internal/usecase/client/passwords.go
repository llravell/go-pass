package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/llravell/go-pass/pkg/encryption"
)

type PasswordsUseCase struct {
	passwordRepo PasswordRepository
}

func NewPasswordsUseCase(
	passwordRepo PasswordRepository,
) *PasswordsUseCase {
	return &PasswordsUseCase{
		passwordRepo: passwordRepo,
	}
}

func (p *PasswordsUseCase) AddNewPassword(
	ctx context.Context,
	key *encryption.Key,
	name, password, meta string,
) error {
	exists, err := p.passwordRepo.PasswordExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return entity.ErrPasswordAlreadyExist
	}

	encryptedPass, err := key.Encrypt(password)
	if err != nil {
		return err
	}

	err = p.passwordRepo.CreateNewPassword(ctx, &entity.Password{
		Name:    name,
		Value:   encryptedPass,
		Meta:    meta,
		Version: 1,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *PasswordsUseCase) GetPasswordByName(
	ctx context.Context,
	name string,
) (*entity.Password, error) {
	return p.passwordRepo.GetPasswordByName(ctx, name)
}

func (p *PasswordsUseCase) UpdatePassword(
	ctx context.Context,
	password *entity.Password,
) error {
	return p.passwordRepo.UpdatePassword(ctx, password)
}

func (p *PasswordsUseCase) GetList(
	ctx context.Context,
) ([]*entity.Password, error) {
	return p.passwordRepo.GetPasswords(ctx)
}
