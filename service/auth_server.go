package service

import (
	"context"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthUserServer is the service for user authentication
type AuthUserServer struct {
	pb.UnimplementedAuthServiceServer

	userStore  UserStore
	jwtManager *JWTManager
}

// NewAuthUserServer creates a new AuthUser server
func NewAuthUserServer(userStore UserStore, jwtManager *JWTManager) *AuthUserServer {
	return &AuthUserServer{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

// Login authenticates a user with the given credentials
func (s *AuthUserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.userStore.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error finding user by username: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating token: %v", err)
	}

	res := &pb.LoginResponse{AccessToken: token}
	return res, nil
}
