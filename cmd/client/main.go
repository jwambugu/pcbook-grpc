package main

import (
	"context"
	"flag"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func createLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := factory.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("searching for laptop with filter %v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}

	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatalf("failed to search laptop: %v", err)
	}

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("failed to receive response: %v", err)
		}

		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberOfCores())
		log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinimumFrequency())
		log.Print("  + ram: ", laptop.GetRam())
		log.Print("  + price: ", laptop.GetPriceUsd())
	}
}

func main() {
	serverAddress := flag.String("server-address", "0.0.0.0:8080", "the server address")
	flag.Parse()

	log.Printf("dialing server %s...", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial server %s: %v", *serverAddress, err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	for i := 1; i <= 10; i++ {
		createLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPriceUsd:     3000,
		MinCpuCores:     4,
		MinCpuFrequency: 2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	searchLaptop(laptopClient, filter)
}
