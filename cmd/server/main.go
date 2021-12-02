package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

const (
	// TODO: read from config
	jwtSecretKey     = "67#$>-,x?`TSZe]\"<B{}&}8}/Gj]b$T>"
	jwtTokenDuration = 15 * time.Minute
)

// unaryInterceptor provides a hook to intercept the execution of a streaming RPC on the server. of a unary RPC on the server
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Printf("[*] unaryInterceptor(_) %v", info.FullMethod)
	return handler(ctx, req)
}

// streamInterceptor provides a hook to intercept the execution of a streaming RPC on the server.
func streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("[*] streamInterceptor(_) %v", info.FullMethod)
	return handler(srv, ss)
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}

func seedUsers(userStore service.UserStore) error {
	if err := createUser(userStore, "admin", "secret", "admin"); err != nil {
		return err
	}

	return createUser(userStore, "user", "secret", "user")
}

func main() {
	port := flag.Int("port", 8080, "server port to listen on")
	flag.Parse()

	log.Printf("starting server on port %d", *port)

	userStore := service.NewInMemoryUserStore()

	if err := seedUsers(userStore); err != nil {
		log.Fatalf("could not seed users: %v", err)
	}

	jwtManager := service.NewJWTManager(jwtSecretKey, jwtTokenDuration)
	authUserServer := service.NewAuthUserServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("storage/public")
	ratingStore := service.NewInMemoryRatingStore()

	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authUserServer)
	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not listen on %s: %v", address, err)
	}

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
