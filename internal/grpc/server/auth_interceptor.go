package server

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type contextKey string

var UserIDContextKey contextKey = "userID"

type JWTParser interface {
	Parse(tokenString string) (*jwt.Token, error)
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	value := ctx.Value(UserIDContextKey)
	id, ok := value.(int)

	return id, ok
}

func getUserIDFromToken(token *jwt.Token) int {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return 0
	}

	id, err := strconv.Atoi(sub)
	if err != nil {
		return 0
	}

	return id
}

func AuthInterceptor(
	jwtParser JWTParser,
) grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		tokenString, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		token, err := jwtParser.Parse(tokenString)
		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid auth token")
		}

		userID := getUserIDFromToken(token)
		ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", userID})

		return context.WithValue(ctx, UserIDContextKey, userID), nil
	})
}
