package client

import (
	"context"

	"github.com/llravell/go-pass/internal/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	headerAuthorize = "authorization"
)

type SessionRepository interface {
	GetSession(ctx context.Context) (*entity.ClientSession, error)
}

func AuthInterceptor(
	sessionRepo SessionRepository,
) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		session, err := sessionRepo.GetSession(ctx)
		if err != nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		if len(session.AuthToken) == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		authorize := "bearer " + session.AuthToken
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headerAuthorize, authorize))

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AuthStreamInterceptor(
	sessionRepo SessionRepository,
) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		session, err := sessionRepo.GetSession(ctx)
		if err != nil {
			return streamer(ctx, desc, cc, method, opts...)
		}

		if len(session.AuthToken) == 0 {
			return streamer(ctx, desc, cc, method, opts...)
		}

		authorize := "bearer " + session.AuthToken
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headerAuthorize, authorize))

		return streamer(ctx, desc, cc, method, opts...)
	}
}
