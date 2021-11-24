package serializer

import (
	"github.com/jwambugu/pcbook-grpc/factory"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/protobuf/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptopOne := factory.NewLaptop()

	err := WriteProtobufToBinaryFile(laptopOne, binaryFile)
	require.NoError(t, err)

	laptopTwo := &pb.Laptop{}

	err = ReadProtobufFromBinaryFile(binaryFile, laptopTwo)

	require.NoError(t, err)
	require.True(t, proto.Equal(laptopOne, laptopTwo))

	err = WriteProtobufToJSONFile(laptopOne, jsonFile)
	require.NoError(t, err)
}
