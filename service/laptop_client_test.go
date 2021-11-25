package service

import (
	"context"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func startLaptopTestServer(t *testing.T) (*LaptopServer, string) {
	laptopServer := NewLaptopServer(NewInMemoryLaptopStore())

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	return laptopServer, listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, address string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, expected *pb.Laptop, actual *pb.Laptop) {
	expectedJSON, err := serializer.ProtobufToJSON(expected)
	require.NoError(t, err)

	actualJSON, err := serializer.ProtobufToJSON(actual)
	require.NoError(t, err)

	require.Equal(t, expectedJSON, actualJSON)
}

func TestLaptopClient_CreateLaptop(t *testing.T) {
	t.Parallel()

	laptopServer, serverAddress := startLaptopTestServer(t)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := factory.NewLaptop()
	expectedLaptopID := laptop.Id

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedLaptopID, res.Id)

	createdLaptop, err := laptopServer.Store.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, createdLaptop)

	requireSameLaptop(t, laptop, createdLaptop)
}
