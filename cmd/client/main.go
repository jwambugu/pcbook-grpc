package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/client"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const refreshDuration = 30 * time.Second

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(factory.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	filter := &pb.Filter{
		MaxPriceUsd:     3000,
		MinCpuCores:     4,
		MinCpuFrequency: 2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := factory.NewLaptop()

	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.jpg")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDS := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := factory.NewLaptop()

		laptopClient.CreateLaptop(laptop)
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

		if err := laptopClient.RateLaptop(laptopIDS, scores); err != nil {
			log.Fatalf("failed to rate laptop: %v", err)
		}
	}
}

func authMethods() map[string]struct{} {
	const laptopServicePath = "/pcbook.LaptopService"

	return map[string]struct{}{
		fmt.Sprintf("%s/CreateLaptop", laptopServicePath): {},
		fmt.Sprintf("%s/UploadImage", laptopServicePath):  {},
		fmt.Sprintf("%s/RateLaptop", laptopServicePath):   {},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA that signed the server's certificate
	// Allows the client to verify authenticity the server's certificate
	file, err := os.Open("certs/ca-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to open server CA pem: %v", err)
	}

	contents, err := io.ReadAll(file)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read server CA pem: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(contents) {
		return nil, fmt.Errorf("failed to append server CA pem to cert pool")
	}

	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	serverAddress := flag.String("server-address", "0.0.0.0:8080", "the server address")
	flag.Parse()

	log.Printf("dialing server %s...", *serverAddress)

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		log.Fatalf("failed to dial server %s: %v", *serverAddress, err)
	}

	authClient := client.NewAuthClient(conn, "admin", "secret")

	fmt.Println(authMethods())

	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatalf("failed to create auth interceptor: %v", err)
	}

	cc, err := grpc.Dial(
		*serverAddress,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)

	if err != nil {
		log.Fatalf("failed to dial server with interceptors %s: %v", *serverAddress, err)
	}

	laptopClient := client.NewLaptopClient(cc)

	testCreateLaptop(laptopClient)
	testUploadImage(laptopClient)

	testRateLaptop(laptopClient)
}
