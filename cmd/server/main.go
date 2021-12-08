package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	// TODO: read from config
	jwtSecretKey     = "67#$>-,x?`TSZe]\"<B{}&}8}/Gj]b$T>"
	jwtTokenDuration = 15 * time.Minute
)

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

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/pcbook.LaptopService"

	return map[string][]string{
		fmt.Sprintf("%s/CreateLaptop", laptopServicePath): {"admin"},
		fmt.Sprintf("%s/UploadImage", laptopServicePath):  {"admin"},
		fmt.Sprintf("%s/RateLaptop", laptopServicePath):   {"admin", "user"},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA that signed the client's certificate
	// Allows the client to verify authenticity the client's certificate
	file, err := os.Open("certs/ca-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to open server CA pem: %v", err)
	}

	contents, err := io.ReadAll(file)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read client CA pem: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(contents) {
		return nil, fmt.Errorf("failed to append client CA pem to cert pool")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("certs/server-cert.pem", "certs/server-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
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

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatalf("could not load TLS credentials: %v", err)
	}

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
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
