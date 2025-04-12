package server

import (
	"context"
	"errors"
	"time"

	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServer

	authUC *usecase.AuthUseCase
	log    *zerolog.Logger
}

func NewAuthServer(
	authUC *usecase.AuthUseCase,
	log *zerolog.Logger,
) *AuthServer {
	return &AuthServer{
		authUC: authUC,
		log:    log,
	}
}

func (s *AuthServer) Register(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	user, err := s.authUC.RegisterUser(ctx, in.GetLogin(), in.GetPassword())

	if err != nil {
		if errors.Is(err, entity.ErrUserConflict) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		s.log.Error().Err(err).Msg("user saving failed")

		return nil, status.Error(codes.Internal, "user saving failed")
	}

	token, err := s.authUC.BuildUserToken(user, 24*time.Hour)
	if err != nil {
		s.log.Error().Err(err).Msg("token issuing failed")

		return nil, status.Error(codes.Internal, "token issuing failed")
	}

	return &pb.AuthResponse{Token: token}, nil
}

func (s *AuthServer) Login(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	user, err := s.authUC.VerifyUser(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		s.log.Error().Err(err).Msg("login failed")

		return nil, status.Error(codes.Internal, "login failed")
	}

	token, err := s.authUC.BuildUserToken(user, 24*time.Hour)
	if err != nil {
		s.log.Error().Err(err).Msg("token issuing failed")

		return nil, status.Error(codes.Internal, "token issuing failed")
	}

	return &pb.AuthResponse{Token: token}, nil
}

// AuthFuncOverride отключает проверку авторизации в интерсепторе AuthServerInterceptor.
func (s *AuthServer) AuthFuncOverride(ctx context.Context, _ string) (context.Context, error) {
	return ctx, nil
}
