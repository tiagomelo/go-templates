package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserIdKey string

type Validator interface {
	ValidateJWTToken(token string) (string, error)
}

type AuthInterceptor struct {
	validator Validator
}

func NewAuthInterceptor(validator Validator) *AuthInterceptor {
	return &AuthInterceptor{
		validator: validator,
	}
}

func (i *AuthInterceptor) UnaryAuthMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	// get metadata object
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// extract token from authorization header
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// validate token and retrieve the userID
	userID, err := i.validator.ValidateJWTToken(token[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// add our user ID to the context, so we can use it in our RPC handler
	ctx = context.WithValue(ctx, UserIdKey("user_id"), userID)

	// call our handler
	return handler(ctx, req)
}
