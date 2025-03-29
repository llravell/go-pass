package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UnexpectedSigningMethodError struct {
	alg string
}

func (e *UnexpectedSigningMethodError) Error() string {
	return "unexpected signing method: " + e.alg
}

type JWTManager struct {
	secret []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
	}
}

func (m *JWTManager) Issue(userID int, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		Subject:   strconv.Itoa(userID),
	})

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *JWTManager) Parse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if alg := token.Method.Alg(); alg != jwt.SigningMethodHS256.Name {
			return nil, &UnexpectedSigningMethodError{alg}
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
