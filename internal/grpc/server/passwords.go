package server

import (
	"context"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PasswordsServer struct {
	pb.UnimplementedPasswordsServer

	passwordsUC *usecase.PasswordsUseCase
	log         *zerolog.Logger
}

func NewPasswordsServer(
	passwordsUC *usecase.PasswordsUseCase,
	log *zerolog.Logger,
) *PasswordsServer {
	return &PasswordsServer{
		passwordsUC: passwordsUC,
		log:         log,
	}
}

func (s *PasswordsServer) Sync(ctx context.Context, in *pb.Password) (*pb.SyncResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.passwordsUC.SyncPassword(ctx, userID, entity.NewPasswordFromPB(in))

	if err == nil {
		return &pb.SyncResponse{Success: true}, nil
	}

	var conflictErr *entity.PasswordConflictError

	if errors.As(err, &conflictErr) {
		return &pb.SyncResponse{
			Success: false,
			Conflict: &pb.Conflict{
				Password: conflictErr.Password().ToPB(),
				Type:     conflictErr.TypePB(),
			},
		}, nil
	}

	return nil, status.Error(codes.Unknown, "sync failed")
}
