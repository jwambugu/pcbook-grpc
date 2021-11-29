package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
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

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, factory.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
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

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("failed to open image file: %v", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatalf("failed to create upload laptop stream: %v", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:      laptopID,
				FileExtension: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatalf("failed to send request: %v - %v", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("failed to read chunk to buffer: %v", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatalf("failed to send chunk: %v - %v", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to complete image upload: %v", err)
	}

	log.Printf("uploaded image with id %s, size - %d", res.GetId(), res.GetSize())
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := factory.NewLaptop()

	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

func rateLaptop(laptopClient pb.LaptopServiceClient, laptopIDS []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("failed to create rate laptop stream: %v", err)
	}

	responseErrorChan := make(chan error)
	// Receive the response from the server
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Printf("rate laptop stream completed")
				responseErrorChan <- nil
				return
			}

			if err != nil {
				log.Printf("failed to receive response: %v", err)
				responseErrorChan <- err
				return
			}

			log.Printf("rate laptop response: %v", res)
		}
	}()

	// Send requests to the server
	for i, id := range laptopIDS {
		req := &pb.RateLaptopRequest{
			LaptopId: id,
			Score:    scores[i],
		}

		if err := stream.Send(req); err != nil {
			return fmt.Errorf("failed to send request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Printf("sent rate laptop request: %v", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("failed to close stream: %v", err)
	}

	err = <-responseErrorChan
	return err
}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	laptopIDS := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := factory.NewLaptop()

		createLaptop(laptopClient, laptop)
		laptopIDS[i] = laptop.GetId()
	}

	scores := make([]float64, n)
	for {
		fmt.Println("[*] rate laptop: (y/n)")
		var input string
		_, _ = fmt.Scan(&input)

		if strings.ToLower(input) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = factory.RandomLaptopScore()
		}

		fmt.Println(scores)

		if err := rateLaptop(laptopClient, laptopIDS, scores); err != nil {
			log.Fatalf("failed to rate laptop: %v", err)
		}
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
	//testUploadImage(laptopClient)

	testRateLaptop(laptopClient)
}
