package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

const maxImageSize = 1 << 20 // 1MB

// LaptopServer is a gRPC server that implements the LaptopServer interface.
type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer

	laptopStore LaptopStore
	imageStore  ImageStore
}

// NewLaptopServer creates a new LaptopServer.
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
	}
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Error(codes.Canceled, "request cancelled by the client")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	default:
		return nil
	}
}

// CreateLaptop is a unary RPC that creates a new laptop.
func (s *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("recieved CreateLaptop(_) request with id - %s: %v", laptop, laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	}

	if len(laptop.Id) == 0 {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate a new laptop ID: %v", err)
		}

		laptop.Id = id.String()
	}

	if err := contextError(ctx); err != nil {
		log.Printf("context error: %v", err)
		return nil, err
	}

	if err := s.laptopStore.Save(laptop); err != nil {
		code := codes.Internal
		if errors.Is(err, ErrRecordExists) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "failed to save laptop: %v", err)
	}

	log.Printf("saved laptop with id - %s: %v", laptop.GetId(), laptop)
	return &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}, nil
}

// SearchLaptop is a server-streaming RPC to search for laptops.
func (s *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("recieved SearchLaptop(_) request with filter - %v", filter)

	err := s.laptopStore.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		res := &pb.SearchLaptopResponse{
			Laptop: laptop,
		}

		if err := stream.Send(res); err != nil {
			return err
		}

		log.Printf("sent SearchLaptop(_) response with laptop - %v", laptop.GetId())
		return nil
	})

	if err != nil {
		return status.Errorf(codes.Internal, "failed to search laptops: %v", err)
	}

	return nil
}

// UploadImage is a client-streaming RPC to upload images.
func (s *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Printf("failed to receive the first request: %v", err)
		return status.Errorf(codes.InvalidArgument, "failed to receive the first request: %v", err)
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageExtension := req.GetInfo().GetFileExtension()

	log.Printf("received UploadImage(_) request with for laptop - %s, extension - %s", laptopID, imageExtension)

	laptop, err := s.laptopStore.Find(laptopID)
	if err != nil {
		log.Printf("failed to find laptop %v", err)
		return status.Errorf(codes.Internal, "failed to find laptop %v", err)
	}

	if laptop == nil {
		log.Printf("laptop %v not found", laptopID)
		return status.Errorf(codes.InvalidArgument, "laptop %v not found", laptopID)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			log.Printf("context error: %v", err)
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("UploadImage(_) no more image data")
			break
		}

		if err != nil {
			log.Printf("UploadImage(_) failed to stream image data: %v", err)
			return status.Errorf(codes.Unknown, "failed to stream image data: %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		imageSize += size

		log.Printf("UploadImage(_) streaming image data - size - %d", size)

		if imageSize > maxImageSize {
			log.Printf("UploadImage(_) image size exceeded the maximum size of %d bytes", maxImageSize)
			return status.Errorf(codes.InvalidArgument, "image size exceeded the maximum size of %d bytes", maxImageSize)
		}

		// Simulate slow writes.
		//time.Sleep(time.Second)

		_, err = imageData.Write(chunk)
		if err != nil {
			log.Printf("UploadImage(_) failed to write image data: %v", err)
			return status.Errorf(codes.Internal, "failed to write image data: %v", err)
		}
	}

	imageID, err := s.imageStore.Save(laptopID, imageExtension, imageData)
	if err != nil {
		log.Printf("UploadImage(_) failed to save image: %v", err)
		return status.Errorf(codes.Internal, "failed to save image: %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	if err := stream.SendAndClose(res); err != nil {
		log.Printf("UploadImage(_) failed to send image response: %v", err)
		return status.Errorf(codes.Internal, "failed to send image data: %v", err)
	}

	log.Printf("UploadImage(_) saved image with id - %s, size - %d", imageID, imageSize)

	return nil
}
