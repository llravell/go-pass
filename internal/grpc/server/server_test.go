package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/llravell/go-pass/internal/grpc/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var errBoom = errors.New("boom")

const defaultUserID = 1

const bufSize = 1024 * 1024

type echoServer struct {
	pb.UnimplementedEchoServer
}

func (s *echoServer) Send(_ context.Context, in *pb.Message) (*pb.Message, error) {
	return in, nil
}

func startGRPCEchoServer(
	t *testing.T,
	unaryInterceptor grpc.UnaryServerInterceptor,
) (pb.EchoClient, func()) {
	t.Helper()

	echo := &echoServer{}

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))
	pb.RegisterEchoServer(server, echo)

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

	return pb.NewEchoClient(conn), closeFn
}

func fakeAuthInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	ctx = context.WithValue(ctx, server.UserIDContextKey, defaultUserID)

	return handler(ctx, req)
}
