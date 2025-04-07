package server_test

import (
	"testing"
	"time"

	"github.com/llravell/go-pass/internal/grpc/server"
	"github.com/llravell/go-pass/pkg/auth"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret")

	client, closeFn := startGRPCEchoServer(
		t,
		server.AuthInterceptor(jwtManager),
	)
	defer closeFn()

	t.Run("interceptor reject request without auth header", func(t *testing.T) {
		_, err := client.Send(t.Context(), &pb.Message{})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("interceptor reject request with invalid auth header", func(t *testing.T) {
		token, err := jwtManager.Issue(1, time.Millisecond)
		require.NoError(t, err)

		md := metadata.Pairs("authorization", "bearer "+token)

		time.Sleep(5 * time.Millisecond)

		_, err = client.Send(metadata.NewOutgoingContext(t.Context(), md), &pb.Message{})

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("interceptor allow request with correct auth header", func(t *testing.T) {
		token, err := jwtManager.Issue(1, time.Hour)
		require.NoError(t, err)

		md := metadata.Pairs("authorization", "bearer "+token)

		_, err = client.Send(metadata.NewOutgoingContext(t.Context(), md), &pb.Message{})
		require.NoError(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)

		assert.Equal(t, codes.OK, st.Code())
	})
}
