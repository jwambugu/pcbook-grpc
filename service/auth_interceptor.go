package service

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

// AuthInterceptor is a server interceptor for authentication and authorization.
type AuthInterceptor struct {
	jwtManager      *JWTManager
	accessibleRoles map[string][]string
}

// NewAuthInterceptor creates a new AuthInterceptor.
func NewAuthInterceptor(jwtManager *JWTManager, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager:      jwtManager,
		accessibleRoles: accessibleRoles,
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context, method string) error {
	roles, ok := i.accessibleRoles[method]
	if !ok {
		// Method is accessible by all users.
		return nil
	}

	ctxMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := ctxMetadata["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing authorization token")
	}

	accessToken := values[0]
	claims, err := i.jwtManager.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid access token provided: %v", err)
	}

	for _, role := range roles {
		if role == claims.Role {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "not authorized to access %q", method)
}

// Unary returns a new unary server interceptor for authentication and authorization of unary RPC calls.
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		log.Printf("[*] unaryInterceptor(_) %v", info.FullMethod)

		// Check if the method is accessible by the user.
		if err := i.authorize(ctx, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a new stream server interceptor for authentication and authorization of stream RPC calls.
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Printf("[*] streamInterceptor(_) %v", info.FullMethod)

		// Check if the method is accessible by the user.
		if err := i.authorize(ss.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
