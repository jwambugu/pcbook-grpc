package client

import (
	"context"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc"
	"time"
)

// AuthClient is a client for the auth service.
type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

// NewAuthClient creates a new AuthClient.
func NewAuthClient(cc *grpc.ClientConn, username string, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)

	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

// Login authenticates the user. On success login, it returns the user's token.
func (c *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: c.username,
		Password: c.password,
	}

	res, err := c.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
