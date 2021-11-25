package service

import (
	"context"
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestLaptopServer_CreateLaptop(t *testing.T) {
	t.Parallel()

	laptopWithNoID := factory.NewLaptop()
	laptopWithNoID.Id = ""

	laptopWithInvalidID := factory.NewLaptop()
	laptopWithInvalidID.Id = "invalid-uuid"

	laptopWithDuplicateID := factory.NewLaptop()
	storeDuplicateID := NewInMemoryLaptopStore()

	err := storeDuplicateID.Save(laptopWithDuplicateID)
	require.NoError(t, err)

	testCases := []struct {
		name   string
		laptop *pb.Laptop
		store  LaptopStore
		code   codes.Code
	}{
		{
			name:   "creates a laptop successfully with id provided",
			laptop: factory.NewLaptop(),
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "fails to create a laptop with invalid id",
			laptop: laptopWithInvalidID,
			store:  NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "fails to create a laptop if the id already exists",
			laptop: laptopWithDuplicateID,
			store:  storeDuplicateID,
			code:   codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := NewLaptopServer(tc.store)
			res, err := server.CreateLaptop(context.Background(), req)

			if tc.code != codes.OK {
				require.Error(t, err)
				require.Nil(t, res)

				statusFromError, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, statusFromError.Code())

				return
			}

			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res)

				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id)
				}
			}
		})
	}
}
