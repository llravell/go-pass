package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/llravell/go-pass/internal/entity"
)

var ErrPasswordDeleteConflict = errors.New("password has been deleted")

type PasswordDiffConflictError struct {
	password *entity.Password
}

func (e *PasswordDiffConflictError) GetConflictedPassword() *entity.Password {
	return e.password
}

func (e *PasswordDiffConflictError) Error() string {
	return fmt.Sprintf("conflicted with %d version", e.password.Version)
}

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
		func(targetPassword *entity.Password) (*entity.Password, error) {
			if targetPassword.Deleted {
				if password.Version > targetPassword.Version {
					return password, nil
				}

				return nil, ErrPasswordDeleteConflict
			}

			if password.Version > targetPassword.Version {
				return password, nil
			}

			return nil, &PasswordDiffConflictError{password: targetPassword}
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
