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

func (uc *PasswordsUseCase) SyncPassword(
	ctx context.Context,
	userID int,
	password *entity.Password,
) error {
	err := uc.repo.UpdateByName(
		ctx,
		userID,
		password.Name,
		func(basePassword *entity.Password) (*entity.Password, error) {
			if basePassword.Deleted {
				if password.Version > basePassword.Version {
					return password, nil
				}

				return nil, entity.NewPasswordDeletedConflictError(basePassword)
			}

			if password.Version > basePassword.Version {
				return password, nil
			}

			return nil, entity.NewPasswordDiffConflictError(basePassword)
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
