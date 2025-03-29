package client

import (
	"context"
	"fmt"

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

func AuthClientInterceptor(
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

		authorize := fmt.Sprintf("bearer %s", session.AuthToken)
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headerAuthorize, authorize))

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
