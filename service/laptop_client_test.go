package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/jwambugu/pcbook-grpc/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func startLaptopTestServer(t *testing.T, laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) string {
	laptopServer := NewLaptopServer(laptopStore, imageStore, ratingStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	return listener.Addr().String()
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

	laptopStore := NewInMemoryLaptopStore()

	serverAddress := startLaptopTestServer(t, laptopStore, nil, nil)
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

	createdLaptop, err := laptopStore.Find(res.Id)
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

	laptopStore := NewInMemoryLaptopStore()
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

		err := laptopStore.Save(laptop)
		require.NoError(t, err)
	}

	serverAddress := startLaptopTestServer(t, laptopStore, nil, nil)
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
		require.Contains(t, expectedIDS, res.Laptop.GetId())
		found += 1
	}

	require.Equal(t, len(expectedIDS), found)
}

func TestLaptopServer_UploadImage(t *testing.T) {
	t.Parallel()

	testImagesFolder := "../tmp"

	laptopStore := NewInMemoryLaptopStore()
	imageStore := NewDiskImageStore(testImagesFolder)

	laptop := factory.NewLaptop()

	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startLaptopTestServer(t, laptopStore, imageStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	imagePath := fmt.Sprintf("%s/laptop.jpg", testImagesFolder)

	file, err := os.Open(imagePath)
	require.NoError(t, err)

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	ext := filepath.Ext(imagePath)

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:      laptop.GetId(),
				FileExtension: ext,
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		size += n

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotEmpty(t, res.GetId())
	require.EqualValues(t, size, res.GetSize())

	uploadedImagePath := fmt.Sprintf("%s/%s%s", testImagesFolder, res.GetId(), ext)
	require.FileExists(t, uploadedImagePath)
	require.NoError(t, os.Remove(uploadedImagePath))
}

func TestLaptopServer_RateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()
	ratingStore := NewInMemoryRatingStore()

	laptop := factory.NewLaptop()

	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startLaptopTestServer(t, laptopStore, nil, ratingStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	stream, err := laptopClient.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averageScores := []float64{8, 7.75, 8.5}

	n := len(scores)

	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			break
		}

		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(idx+1), res.GetRatingsCount())
		require.Equal(t, averageScores[idx], res.GetAverageScore())
	}
}
