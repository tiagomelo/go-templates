package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestIssueJWTToken(t *testing.T) {
	testCases := []struct {
		name                  string
		userID                string
		mockTokenSignedString func(token *jwt.Token, secret []byte) (string, error)
		expectedError         error
	}{
		{
			name:   "happy path",
			userID: "user-id",
			mockTokenSignedString: func(token *jwt.Token, secret []byte) (string, error) {
				return "token", nil
			},
		},
		{
			name:   "token signed string error",
			userID: "user-id",
			mockTokenSignedString: func(token *jwt.Token, secret []byte) (string, error) {
				return "", errors.New("sign token error")
			},
			expectedError: errors.Wrap(errors.New("sign token error"), "signing JWT token"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenSignedString = tc.mockTokenSignedString
			s := NewService("secret")
			token, err := s.IssueJWTToken(tc.userID)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateJWTToken(t *testing.T) {
	testCases := []struct {
		name          string
		mockJwtParse  func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error)
		expectedError error
	}{
		{
			name: "happy path",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return &jwt.Token{
					Valid:  true,
					Claims: jwt.MapClaims{"sub": "user-id"},
					Method: jwt.SigningMethodHS256,
				}, nil
			},
		},
		{
			name: "invalid signing method",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				token := &jwt.Token{
					Valid:  true,
					Claims: jwt.MapClaims{"sub": "user-id"},
					Method: nil,
				}
				_, err := keyFunc(token)
				if err != nil {
					return nil, err
				}
				return token, nil
			},
			expectedError: errors.Wrap(ErrInvalidJWTToken, "invalid signing method: <nil>"),
		},
		{
			name: "error when parsing token",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return nil, errors.New("parse token error")
			},
			expectedError: errors.Wrap(ErrInvalidJWTToken, "parse token error"),
		},
		{
			name: "invalid claims map",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return &jwt.Token{
					Valid:  true,
					Claims: nil,
					Method: jwt.SigningMethodHS256,
				}, nil
			},
			expectedError: ErrInvalidJWTToken,
		},
		{
			name: "invalid claims",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return &jwt.Token{
					Valid:  false,
					Claims: jwt.MapClaims{"sub": 123},
					Method: jwt.SigningMethodHS256,
				}, nil
			},
			expectedError: ErrInvalidJWTToken,
		},
		{
			name: "sub is not a string",
			mockJwtParse: func(tokenString string, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return &jwt.Token{
					Valid:  true,
					Claims: jwt.MapClaims{"sub": 123},
					Method: jwt.SigningMethodHS256,
				}, nil
			},
			expectedError: errors.Wrap(ErrInvalidJWTToken, "extracting user id from claims"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwtParse = tc.mockJwtParse
			s := NewService("secret")
			token, err := s.ValidateJWTToken("token")
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotEmpty(t, token)
			}
		})
	}
}
