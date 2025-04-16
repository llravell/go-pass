package server

import (
	"context"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type PasswordsUseCase struct {
	repo PasswordsRepository
}

func NewPasswordsUseCase(repo PasswordsRepository) *PasswordsUseCase {
	return &PasswordsUseCase{
		repo: repo,
	}
}

func (uc *PasswordsUseCase) AddNewPassword(
	ctx context.Context,
	userID int,
	password *entity.Password,
) error {
	return uc.repo.AddNewPassword(ctx, userID, password)
}

func (uc *PasswordsUseCase) DeletePasswordByName(
	ctx context.Context,
	userID int,
	name string,
) error {
	return uc.repo.DeletePasswordByName(ctx, userID, name)
}

func (uc *PasswordsUseCase) GetPasswords(
	ctx context.Context,
	userID int,
) ([]*entity.Password, error) {
	return uc.repo.GetPasswords(ctx, userID)
}

func (uc *PasswordsUseCase) SyncPassword(
	ctx context.Context,
	userID int,
	password *entity.Password,
) error {
	err := uc.repo.UpdateByName(
		ctx,
		userID,
		password.Name,
		func(actualPassword *entity.Password) (*entity.Password, error) {
			return entity.ChooseMostActuralEntity[entity.Password](actualPassword, password)
		},
	)
	if err != nil {
		if !errors.Is(err, entity.ErrPasswordDoesNotExist) {
			return err
		}

		err = uc.AddNewPassword(ctx, userID, password)
		if err != nil {
			return err
		}
	}

	return nil
}
