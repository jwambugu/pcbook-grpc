package main

import (
	"flag"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 8080, "server port to listen on")
	flag.Parse()

	log.Printf("starting server on port %d", *port)

	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not listen on %s: %v", address, err)
	}

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
