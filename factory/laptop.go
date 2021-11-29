package factory

import (
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewKeyboard returns a new sample keyboard layout
func NewKeyboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:    randomKeyboardLayout(),
		IsBacklit: randomBool(),
	}
}

// NewCPU returns a new sample CPU
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	numberOfCores := randomInt(2, 8)
	numberOfThreads := randomInt(numberOfCores, 12)
	minimumFrequency := randomFloat64(2.0, 3.5)
	maximumFrequency := randomFloat64(minimumFrequency, 5.0)

	cpu := &pb.CPU{
		Brand:            brand,
		Name:             randomCPUName(brand),
		NumberOfCores:    uint32(numberOfCores),
		NumberOfThreads:  uint32(numberOfThreads),
		MinimumFrequency: minimumFrequency,
		MaximumFrequency: maximumFrequency,
	}

	return cpu
}

// NewGPU returns a new sample GPU
func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minimumFrequency := randomFloat64(1.0, 1.5)
	maximumFrequency := randomFloat64(minimumFrequency, 2.0)

	gpu := &pb.GPU{
		Brand:            brand,
		Name:             name,
		MinimumFrequency: minimumFrequency,
		MaximumFrequency: maximumFrequency,
		Memory: &pb.Memory{
			Value: uint64(randomInt(2, 6)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	return gpu
}

// NewRAM returns a new sample RAM
func NewRAM() *pb.Memory {
	return &pb.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit:  pb.Memory_GIGABYTE,
	}
}

// NewSSD returns a new sample SSD
func NewSSD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(256, 1024)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
}

// NewHDD returns a new sample HDD
func NewHDD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pb.Memory_TERABYTE,
		},
	}
}

// NewScreen returns a new sample Screen
func NewScreen() *pb.Screen {
	screen := &pb.Screen{
		SizeInches:   randomFloat32(13, 17),
		Resolution:   randomScreenResolution(),
		Panel:        randomScreenPanel(),
		IsMultiTouch: randomBool(),
	}

	return screen
}

// NewLaptop returns a new sample Laptop
func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)

	return &pb.Laptop{
		Id:       randomUUID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Gpus:     []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd:    randomFloat64(1500, 3500),
		ReleaseYear: uint32(randomInt(2015, 2019)),
		UpdatedAt:   timestamppb.Now(),
	}
}

// RandomLaptopScore returns a random laptop score
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
