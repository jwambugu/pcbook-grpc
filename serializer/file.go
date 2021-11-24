package serializer

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
)

// WriteProtobufToBinaryFile writes protocol buffer message to binary file
func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("cannot serialize proto message to binary: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("cannot write binary data to file: %w", err)
	}

	return nil
}

// ReadProtobufFromBinaryFile reads protocol buffer message from binary file
func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open binary file: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read binary file: %w", err)
	}

	if err := proto.Unmarshal(data, message); err != nil {
		return fmt.Errorf("cannot serialize binary to proto message: %w", err)
	}

	return nil
}

// WriteProtobufToJSONFile writes protocol buffer message to JSON file
func WriteProtobufToJSONFile(message proto.Message, filename string) error {
	data, err := ProtobufToJSON(message)
	if err != nil {
		return fmt.Errorf("cannot serialize proto message to JSON: %w", err)
	}

	if err := os.WriteFile(filename, []byte(data), 0644); err != nil {
		return fmt.Errorf("cannot write JSON data to file: %w", err)
	}

	return nil
}
