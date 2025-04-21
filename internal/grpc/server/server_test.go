package server_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"testing"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/llravell/go-pass/internal/entity"
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

type fileMatcher struct {
	name   string
	bucket string
}

//nolint:unparam
func newFileMatcher(name, bucket string) *fileMatcher {
	return &fileMatcher{
		name:   name,
		bucket: bucket,
	}
}

func (m *fileMatcher) Matches(x any) bool {
	file, ok := x.(*entity.File)
	if !ok {
		return false
	}

	return file.Name == m.name && file.MinioBucket == m.bucket
}

func (m *fileMatcher) String() string {
	return fmt.Sprintf("match that file has name=\"%s\" and bucket=\"%s\"", m.name, m.bucket)
}

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

func fakeAuthStreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := context.WithValue(ss.Context(), server.UserIDContextKey, defaultUserID)

	wrapped := middleware.WrapServerStream(ss)
	wrapped.WrappedContext = ctx

	return handler(srv, wrapped)
}
