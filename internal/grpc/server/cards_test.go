package server_test

import (
	"context"
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

func startGRPCCardsServer(
	t *testing.T,
	cardsRepo usecase.CardsRepository,
) (pb.CardsClient, func()) {
	t.Helper()

	logger := zerolog.Nop()
	cardsUsecase := usecase.NewCardsUseCase(cardsRepo)
	cardsServer := server.NewCardsServer(cardsUsecase, &logger)

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer(grpc.UnaryInterceptor(fakeAuthInterceptor))
	pb.RegisterCardsServer(server, cardsServer)

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

	return pb.NewCardsClient(conn), closeFn
}

func TestCardsServer_GetList(t *testing.T) {
	cardsRepo := mocks.NewMockCardsRepository(gomock.NewController(t))

	client, closeFn := startGRPCCardsServer(t, cardsRepo)

	defer closeFn()

	t.Run("return cards list", func(t *testing.T) {
		cardsRepo.EXPECT().
			GetCards(gomock.Any(), defaultUserID).
			Return([]*entity.Card{
				{Name: "a"},
				{Name: "b"},
			}, nil)

		resp, err := client.GetList(t.Context(), &emptypb.Empty{})
		require.NoError(t, err)

		cards := resp.GetCards()
		assert.Equal(t, 2, len(cards))

		first, second := cards[0], cards[1]

		assert.Equal(t, "a", first.GetName())
		assert.Equal(t, "b", second.GetName())
	})

	t.Run("cards fetching error", func(t *testing.T) {
		cardsRepo.EXPECT().
			GetCards(gomock.Any(), defaultUserID).
			Return(nil, errBoom)

		_, err := client.GetList(t.Context(), &emptypb.Empty{})

		status, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, status.Code())
	})
}

func TestCardsServer_Delete(t *testing.T) {
	cardsRepo := mocks.NewMockCardsRepository(gomock.NewController(t))

	client, closeFn := startGRPCCardsServer(t, cardsRepo)

	defer closeFn()

	t.Run("delete card", func(t *testing.T) {
		cardsRepo.EXPECT().
			DeleteCardByName(gomock.Any(), defaultUserID, "a").
			Return(nil)

		_, err := client.Delete(t.Context(), &pb.CardDeleteRequest{Name: "a"})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.OK, st.Code())
	})

	t.Run("deleting error", func(t *testing.T) {
		cardsRepo.EXPECT().
			DeleteCardByName(gomock.Any(), defaultUserID, "a").
			Return(errBoom)

		_, err := client.Delete(t.Context(), &pb.CardDeleteRequest{Name: "a"})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Internal, st.Code())
	})
}
