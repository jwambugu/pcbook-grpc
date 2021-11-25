package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// LaptopServer is a gRPC server that implements the LaptopServer interface.
type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer

	Store LaptopStore
}

// NewLaptopServer creates a new LaptopServer.
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
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

	if ctx.Err() == context.Canceled {
		log.Println("CreateLaptop(_) RPC request cancelled")
		return nil, status.Error(codes.Canceled, "request cancelled by the client")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("CreateLaptop(_) RPC request deadline timed out")
		return nil, status.Error(codes.DeadlineExceeded, "request deadline timed out")
	}

	if err := s.Store.Save(laptop); err != nil {
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

	err := s.Store.Search(filter, func(laptop *pb.Laptop) error {
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
