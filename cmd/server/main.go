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

func main() {
	port := flag.Int("port", 8080, "server port to listen on")
	flag.Parse()

	log.Printf("starting server on port %d", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("storage/public")
	ratingStore := service.NewInMemoryRatingStore()

	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
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
