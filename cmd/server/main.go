package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// TODO: read from config
	jwtSecretKey     = "67#$>-,x?`TSZe]\"<B{}&}8}/Gj]b$T>"
	jwtTokenDuration = 15 * time.Minute

	serverCertFile = "certs/server-cert.pem"
	serverKeyFile  = "certs/server-key.pem"
)

type runServerOpts struct {
	listener       net.Listener
	authUserServer pb.AuthServiceServer
	laptopServer   pb.LaptopServiceServer
	jwtManager     *service.JWTManager
	enableTLS      bool
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
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
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

// runGRPCServer runs the gRPC server with the given options
func runGRPCServer(opts runServerOpts) error {
	interceptor := service.NewAuthInterceptor(opts.jwtManager, accessibleRoles())

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if opts.enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("could not load TLS credentials: %v", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterLaptopServiceServer(grpcServer, opts.laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, opts.authUserServer)
	reflection.Register(grpcServer)

	log.Printf("Starting GRPC server on %s, TLS = %t", opts.listener.Addr().String(), opts.enableTLS)
	return grpcServer.Serve(opts.listener)
}

// runRESTServer runs the REST server with the given options
func runRESTServer(opts runServerOpts) error {
	mux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pb.RegisterAuthServiceHandlerServer(ctx, mux, opts.authUserServer); err != nil {
		return err
	}

	if err := pb.RegisterLaptopServiceHandlerServer(ctx, mux, opts.laptopServer); err != nil {
		return err
	}

	log.Printf("Starting REST server on %s, TLS = %t", opts.listener.Addr().String(), opts.enableTLS)

	if opts.enableTLS {
		return http.ServeTLS(opts.listener, mux, serverCertFile, serverKeyFile)
	}

	return http.Serve(opts.listener, mux)
}

func main() {
	port := flag.Int("port", 8080, "server port to listen on")
	enableTLS := flag.Bool("enable-tls", false, "enables TLS")
	serverType := flag.String("server-type", "grpc", "type of server to run -  (grpc/rest)")
	flag.Parse()

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

	address := fmt.Sprintf("0.0.0.0:%d", *port)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not listen on %s: %v", address, err)
	}

	opts := runServerOpts{
		listener:       listen,
		authUserServer: authUserServer,
		laptopServer:   laptopServer,
		jwtManager:     jwtManager,
		enableTLS:      *enableTLS,
	}

	if *serverType == "grpc" {
		err = runGRPCServer(opts)
	} else {
		err = runRESTServer(opts)
	}

	if err != nil {
		log.Fatalf("could not run %s server: %v", *serverType, err)
	}
}
