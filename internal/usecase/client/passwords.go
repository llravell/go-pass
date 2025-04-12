package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"golang.org/x/sync/errgroup"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type PasswordsUseCase struct {
	passwordsRepo   PasswordsRepository
	passwordsClient pb.PasswordsClient
}

func NewPasswordsUseCase(
	passwordsRepo PasswordsRepository,
	passwordsClient pb.PasswordsClient,
) *PasswordsUseCase {
	return &PasswordsUseCase{
		passwordsRepo:   passwordsRepo,
		passwordsClient: passwordsClient,
	}
}

func (p *PasswordsUseCase) AddNewPassword(
	ctx context.Context,
	password entity.Password,
) error {
	password.Version = 1

	created, err := p.passwordsRepo.CreateNewPassword(ctx, &password)
	if err != nil {
		return err
	}

	if !created {
		return nil
	}

	response, err := p.passwordsClient.Sync(ctx, password.ToPB())
	if err != nil {
		return err
	}

	if !response.GetSuccess() {
		return entity.NewPasswordConflictErrorFromPB(&password, response.GetConflict())
	}

	return nil
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

	return entity.NewPasswordConflictErrorFromPB(password, response.GetConflict())
}

func (p *PasswordsUseCase) UpdatePasswordLocal(
	ctx context.Context,
	password *entity.Password,
) error {
	return p.passwordsRepo.UpdatePassword(ctx, password)
}

func (p *PasswordsUseCase) GetList(
	ctx context.Context,
) ([]*entity.Password, error) {
	return p.passwordsRepo.GetPasswords(ctx)
}

func (p *PasswordsUseCase) DeletePasswordLocal(
	ctx context.Context,
	name string,
) error {
	return p.passwordsRepo.DeletePasswordHard(ctx, name)
}

func (p *PasswordsUseCase) AddPasswordsLocal(
	ctx context.Context,
	passwords []*entity.Password,
) error {
	return p.passwordsRepo.CreatePasswordsMultiple(ctx, passwords)
}

func (p *PasswordsUseCase) DeletePasswordByName(
	ctx context.Context,
	name string,
) error {
	exists, err := p.passwordsRepo.PasswordExists(ctx, name)
	if err != nil {
		return err
	}

	if !exists {
		return entity.ErrPasswordDoesNotExist
	}

	_, err = p.passwordsClient.Delete(ctx, &pb.PasswordDeleteRequest{Name: name})
	if err != nil {
		return p.passwordsRepo.DeletePasswordSoft(ctx, name)
	}

	return p.passwordsRepo.DeletePasswordHard(ctx, name)
}

func (p *PasswordsUseCase) GetUpdates(
	ctx context.Context,
) (*entity.PasswordsUpdates, error) {
	localPasswords, serverPasswords, err := p.fetchLocalAndServerPasswords(ctx)
	if err != nil {
		return nil, err
	}

	updates := &entity.PasswordsUpdates{
		ToAdd:    make([]*entity.Password, 0, len(serverPasswords)),
		ToUpdate: make([]*entity.Password, 0, len(serverPasswords)),
		ToSync:   make([]*entity.Password, 0, len(localPasswords)),
	}

	for name, serverPass := range serverPasswords {
		localPass, ok := localPasswords[name]
		if !ok {
			updates.ToAdd = append(updates.ToAdd, serverPass)

			continue
		}

		if localPass.Version < serverPass.Version {
			updates.ToUpdate = append(updates.ToUpdate, serverPass)

			continue
		}

		if localPass.Version > serverPass.Version || !localPass.Equal(serverPass) {
			updates.ToSync = append(updates.ToSync, localPass)
		}
	}

	for name, localPass := range localPasswords {
		_, ok := serverPasswords[name]
		if !ok {
			updates.ToSync = append(updates.ToSync, localPass)
		}
	}

	return updates, nil
}

func (p *PasswordsUseCase) fetchLocalAndServerPasswords(
	ctx context.Context,
) (map[string]*entity.Password, map[string]*entity.Password, error) {
	localPasswords := make(map[string]*entity.Password, 0)
	serverPasswords := make(map[string]*entity.Password, 0)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		passwords, err := p.passwordsRepo.GetPasswords(ctx)
		if err != nil {
			return err
		}

		for _, password := range passwords {
			localPasswords[password.Name] = password
		}

		return nil
	})

	group.Go(func() error {
		response, err := p.passwordsClient.GetList(ctx, &emptypb.Empty{})
		if err != nil {
			return err
		}

		for _, pass := range response.GetPasswords() {
			serverPasswords[pass.GetName()] = entity.NewPasswordFromPB(pass)
		}

		return nil
	})

	err := group.Wait()
	if err != nil {
		return nil, nil, err
	}

	return localPasswords, serverPasswords, nil
}
