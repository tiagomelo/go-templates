package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const fifteenMinutes = time.Minute * 15

var (
	// For ease of unit testing.
	tokenSignedString = func(token *jwt.Token, secret []byte) (string, error) {
		return token.SignedString(secret)
	}
	// For ease of unit testing.
	jwtParse = func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		return jwt.Parse(tokenString, keyFunc, options...)
	}
	ErrInvalidJWTToken = errors.New("invalid JWT token")
)

// service provides JWT token related operations.
type service struct {
	secret []byte
}

// NewService creates a new auth service.
func NewService(secret string) *service {
	return &service{secret: []byte(secret)}
}

// IssueJWTToken issues a new JWT token for the given user ID.
func (s *service) IssueJWTToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iss": time.Now().Unix(),
		"exp": time.Now().Add(fifteenMinutes).Unix(),
	})
	signed, err := tokenSignedString(token, s.secret)
	if err != nil {
		return "", errors.Wrap(err, "signing JWT token")
	}
	return signed, nil
}

// ValidateJWTToken validates the given JWT token and returns the user ID.
func (s *service) ValidateJWTToken(token string) (string, error) {
	t, err := jwtParse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return "", errors.Wrap(ErrInvalidJWTToken, err.Error())
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", errors.Wrap(ErrInvalidJWTToken, "extracting user id from claims")
		}
		return userID, nil
	}
	return "", ErrInvalidJWTToken
}
