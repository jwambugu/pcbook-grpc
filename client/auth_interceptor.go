package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

// AuthInterceptor is a client-side interceptor for authentication.
type AuthInterceptor struct {
	authClient  *AuthClient
	authMethods map[string]struct{}
	accessToken string
}

func (i *AuthInterceptor) refreshToken() error {
	accessToken, err := i.authClient.Login()
	if err != nil {
		return err
	}

	i.accessToken = accessToken
	log.Printf("access token refreshed: %v", accessToken)

	return nil
}

func (i *AuthInterceptor) scheduleRefreshToken(duration time.Duration) error {
	if err := i.refreshToken(); err != nil {
		return err
	}

	go func() {
		wait := duration

		for {
			time.Sleep(wait)
			if err := i.refreshToken(); err != nil {
				log.Printf("failed to refresh access token: %v", err)
				wait = time.Second
				continue
			}

			wait = duration
		}
	}()
	return nil
}

func (i *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", i.accessToken)
}

// Unary returns a new unary client interceptor for authentication.
func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log.Printf("[*] Unary(_) %v", method)

		if i.authMethods[method] == struct{}{} {
			return invoker(i.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// Stream returns a new streaming client interceptor for authentication.
func (i *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		log.Printf("[*] Stream(_) %v", method)

		if i.authMethods[method] == struct{}{} {
			return streamer(i.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// NewAuthInterceptor creates a new AuthInterceptor.
func NewAuthInterceptor(
	authClient *AuthClient, authMethods map[string]struct{}, refreshDuration time.Duration,
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
	}

	if err := interceptor.scheduleRefreshToken(refreshDuration); err != nil {
		return interceptor, err
	}

	return interceptor, nil
}
