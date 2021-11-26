package service

import (
	"context"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"
	"net"
	"testing"
)

func startLaptopTestServer(t *testing.T, store LaptopStore) (*LaptopServer, string) {
	laptopServer := NewLaptopServer(store)

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

	laptopServer, serverAddress := startLaptopTestServer(t, NewInMemoryLaptopStore())
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

func TestLaptopServer_SearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd:     2000,
		MinCpuCores:     4,
		MinCpuFrequency: 2.2,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	store := NewInMemoryLaptopStore()
	expectedIDS := make(map[string]struct{})

	for i := 0; i < 6; i++ {
		laptop := factory.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberOfCores = 2
		case 2:
			laptop.Cpu.MinimumFrequency = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberOfCores = 4
			laptop.Cpu.MinimumFrequency = 2.5
			laptop.Cpu.MaximumFrequency = 4.5
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIDS[laptop.Id] = struct{}{}
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberOfCores = 6
			laptop.Cpu.MinimumFrequency = 2.8
			laptop.Cpu.MaximumFrequency = 5.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIDS[laptop.Id] = struct{}{}
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

	_, serverAddress := startLaptopTestServer(t, store)
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIDS, res.Laptop.Id)
		found += 1
	}

	require.Equal(t, len(expectedIDS), found)
}
