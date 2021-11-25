package main

import (
	"context"
	"flag"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

func main() {
	serverAddress := flag.String("server-address", "0.0.0.0:8080", "the server address")
	flag.Parse()

	log.Printf("dialing server %s...", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial server %s: %v", *serverAddress, err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	laptop := factory.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)

	if err != nil {
		statusFromError, ok := status.FromError(err)
		if ok && statusFromError.Code() == codes.AlreadyExists {
			log.Printf("laptop with id %s already exists", laptop.Id)
			return
		}

		log.Fatalf("failed to create laptop: %v", err)
	}

	log.Printf("created laptop with id %s", res.Id)
}
