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
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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

func (s *PasswordsServer) Sync(ctx context.Context, in *pb.Password) (*pb.PasswordSyncResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.passwordsUC.SyncPassword(ctx, userID, entity.NewPasswordFromPB(in))

	if err == nil {
		return &pb.PasswordSyncResponse{Success: true}, nil
	}

	var conflictErr *entity.PasswordConflictError

	if errors.As(err, &conflictErr) {
		s.log.Info().
			Str("conflict_type", string(conflictErr.Type())).
			Msg("sync conflict")

		return &pb.PasswordSyncResponse{
			Success: false,
			Conflict: &pb.Conflict{
				Password: conflictErr.Actual().ToPB(),
				Type:     conflictErr.TypePB(),
			},
		}, nil
	}

	s.log.Error().Err(err).Msg("sync failed")

	return nil, status.Error(codes.Unknown, "sync failed")
}

func (s *PasswordsServer) Delete(ctx context.Context, in *pb.PasswordDeleteRequest) (*emptypb.Empty, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.passwordsUC.DeletePasswordByName(ctx, userID, in.GetName())
	if err != nil {
		s.log.Error().Err(err).Msg("password deleting failed")

		return nil, status.Error(codes.Unknown, "deleting failed")
	}

	return &emptypb.Empty{}, nil
}

func (s *PasswordsServer) GetList(ctx context.Context, _ *emptypb.Empty) (*pb.PasswordGetListResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	passwords, err := s.passwordsUC.GetList(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("password deleting failed")

		return nil, status.Error(codes.Unknown, "deleting failed")
	}

	response := &pb.PasswordGetListResponse{
		Passwords: make([]*pb.Password, 0, len(passwords)),
	}

	for _, password := range passwords {
		response.Passwords = append(response.Passwords, password.ToPB())
	}

	return response, nil
}
