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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func startGRPCPasswordsServer(
	t *testing.T,
	passwordsRepo usecase.PasswordsRepository,
) (pb.PasswordsClient, func()) {
	t.Helper()

	logger := zerolog.Nop()
	passwordsUsecase := usecase.NewPasswordsUseCase(passwordsRepo)
	passwordsServer := server.NewPasswordsServer(passwordsUsecase, &logger)

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer(grpc.UnaryInterceptor(fakeAuthInterceptor))
	pb.RegisterPasswordsServer(server, passwordsServer)

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

	return pb.NewPasswordsClient(conn), closeFn
}

func TestPasswordsServer_GetList(t *testing.T) {
	passwordsRepo := mocks.NewMockPasswordsRepository(gomock.NewController(t))

	client, closeFn := startGRPCPasswordsServer(t, passwordsRepo)

	defer closeFn()

	t.Run("return passwords list", func(t *testing.T) {
		passwordsRepo.EXPECT().
			GetPasswords(gomock.Any(), defaultUserID).
			Return([]*entity.Password{
				&entity.Password{Name: "a"},
				&entity.Password{Name: "b"},
			}, nil)

		resp, err := client.GetList(t.Context(), &emptypb.Empty{})
		require.NoError(t, err)

		passwords := resp.GetPasswords()
		assert.Equal(t, 2, len(passwords))

		first, second := passwords[0], passwords[1]

		assert.Equal(t, "a", first.GetName())
		assert.Equal(t, "b", second.GetName())
	})

	t.Run("passwords fetching error", func(t *testing.T) {
		passwordsRepo.EXPECT().
			GetPasswords(gomock.Any(), defaultUserID).
			Return(nil, errors.New("Boom!"))

		_, err := client.GetList(t.Context(), &emptypb.Empty{})

		status, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Unknown, status.Code())
	})
}

func TestPasswordsServer_Delete(t *testing.T) {
	passwordsRepo := mocks.NewMockPasswordsRepository(gomock.NewController(t))

	client, closeFn := startGRPCPasswordsServer(t, passwordsRepo)

	defer closeFn()

	t.Run("delete password", func(t *testing.T) {
		passwordsRepo.EXPECT().
			DeletePasswordByName(gomock.Any(), defaultUserID, "a").
			Return(nil)

		_, err := client.Delete(t.Context(), &pb.PasswordDeleteRequest{Name: "a"})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.OK, st.Code())
	})

	t.Run("deleting error", func(t *testing.T) {
		passwordsRepo.EXPECT().
			DeletePasswordByName(gomock.Any(), defaultUserID, "a").
			Return(errors.New("Boom!"))

		_, err := client.Delete(t.Context(), &pb.PasswordDeleteRequest{Name: "a"})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Unknown, st.Code())
	})
}
