package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/llravell/go-pass/internal/grpc/server"
	"github.com/llravell/go-pass/internal/mocks"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func encryptUserPassword(t *testing.T, raw string) string {
	t.Helper()

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	require.NoError(t, err)

	return string(passwordBytes)
}

func startGRPCAuthServer(
	t *testing.T,
	userRepo usecase.UserRepository,
	jwtIssuer usecase.JWTIssuer,
) (pb.AuthClient, func()) {
	t.Helper()

	logger := zerolog.Nop()
	authUsecase := usecase.NewAuthUseCase(userRepo, jwtIssuer)
	authServer := server.NewAuthServer(authUsecase, &logger)

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterAuthServer(server, authServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	closeFn := func() {
		conn.Close()
		server.Stop()
	}

	return pb.NewAuthClient(conn), closeFn
}

func TestAuthServer_Register(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(gomock.NewController(t))
	jwtIssuer := mocks.NewMockJWTIssuer(gomock.NewController(t))

	client, closeFn := startGRPCAuthServer(t, userRepo, jwtIssuer)

	defer closeFn()

	t.Run("register new user", func(t *testing.T) {
		userRepo.EXPECT().
			StoreUser(gomock.Any(), "login", gomock.Any()).
			Return(&entity.User{ID: 1, Login: "login"}, nil)

		jwtIssuer.EXPECT().
			Issue(1, gomock.Any()).
			Return("token", nil)

		response, err := client.Register(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})
		require.NoError(t, err)

		assert.Equal(t, "token", response.Token)
	})

	t.Run("already exists error", func(t *testing.T) {
		userRepo.EXPECT().
			StoreUser(gomock.Any(), "login", gomock.Any()).
			Return(nil, entity.ErrUserConflict)

		_, err := client.Register(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("user storing error", func(t *testing.T) {
		userRepo.EXPECT().
			StoreUser(gomock.Any(), "login", gomock.Any()).
			Return(nil, errors.New("Boom!"))

		_, err := client.Register(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("token issuing error", func(t *testing.T) {
		userRepo.EXPECT().
			StoreUser(gomock.Any(), "login", gomock.Any()).
			Return(&entity.User{ID: 1, Login: "login"}, nil)

		jwtIssuer.EXPECT().
			Issue(1, gomock.Any()).
			Return("", errors.New("Boom!"))

		_, err := client.Register(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestAuthServer_Login(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(gomock.NewController(t))
	jwtIssuer := mocks.NewMockJWTIssuer(gomock.NewController(t))

	client, closeFn := startGRPCAuthServer(t, userRepo, jwtIssuer)

	defer closeFn()

	t.Run("login user", func(t *testing.T) {
		userRepo.EXPECT().
			FindUserByLogin(gomock.Any(), "login").
			Return(&entity.User{ID: 1, Login: "login", Password: encryptUserPassword(t, "pass")}, nil)

		jwtIssuer.EXPECT().
			Issue(1, gomock.Any()).
			Return("token", nil)

		response, err := client.Login(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})
		require.NoError(t, err)

		assert.Equal(t, "token", response.Token)
	})

	t.Run("user fetching error", func(t *testing.T) {
		userRepo.EXPECT().
			FindUserByLogin(gomock.Any(), "login").
			Return(nil, errors.New("Boom!"))

		_, err := client.Login(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("token issuing error", func(t *testing.T) {
		userRepo.EXPECT().
			FindUserByLogin(gomock.Any(), "login").
			Return(&entity.User{ID: 1, Login: "login", Password: encryptUserPassword(t, "pass")}, nil)

		jwtIssuer.EXPECT().
			Issue(1, gomock.Any()).
			Return("", errors.New("Boom!"))

		_, err := client.Login(t.Context(), &pb.AuthRequest{
			Login:    "login",
			Password: "pass",
		})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, st.Code())
	})
}
