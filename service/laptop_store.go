package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"log"
	"sync"
)

// ErrRecordExists is an error that is returned when a record already exists
var ErrRecordExists = errors.New("record already exists")

// LaptopStore is an interface for storing laptops
type LaptopStore interface {
	// Save saves a laptop in the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by its id
	Find(id string) (*pb.Laptop, error)

	// Search finds laptops by their properties using a filter, returns one by one laptop via the found function
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// InMemoryLaptopStore is an in-memory implementation of a LaptopStore
type InMemoryLaptopStore struct {
	mutext sync.RWMutex
	data   map[string]*pb.Laptop
}

// NewInMemoryLaptopStore returns a new instance of an InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// matchesFilter returns true if the laptop matches the filter
func matchesFilter(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberOfCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMaximumFrequency() < filter.GetMinCpuFrequency() {
		return false
	}

	if toBits(laptop.GetRam()) < toBits(filter.GetMinRam()) {
		return false
	}

	return true
}

// toBits converts memory unit to bits
func toBits(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3
	case pb.Memory_MEGABYTE:
		return value << 23 // 1024 * 1024 * 8 = 2^10 * 2^10 * 2^3
	case pb.Memory_GIGABYTE:
		return value << 33 // 1024 * 1024 * 1024 * 8 = 2^10 * 2^10 * 2^10 * 2^3
	case pb.Memory_TERABYTE:
		return value << 43 // 1024 * 1024 * 1024 * 1024 * 8 = 2^10 * 2^10 * 2^10 * 2^10 * 2^3
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	l := &pb.Laptop{}

	if err := copier.Copy(l, laptop); err != nil {
		return nil, fmt.Errorf("error copying laptop: %v", err)
	}
	return l, nil
}

// Save saves a laptop in the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutext.Lock()
	defer store.mutext.Unlock()

	if _, ok := store.data[laptop.Id]; ok {
		return ErrRecordExists
	}

	newLaptop, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[laptop.Id] = newLaptop
	return nil
}

// Find finds a laptop by its id
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutext.RLock()
	defer store.mutext.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)
}

// Search finds laptops by their properties using a filter, returns one by one laptop via the found function
func (store *InMemoryLaptopStore) Search(
	ctx context.Context, filter *pb.Filter,
	match func(laptop *pb.Laptop) error,
) error {
	store.mutext.RLock()
	defer store.mutext.RUnlock()

	for _, laptop := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Printf("error searching laptop: %s context cancelled: %v", laptop.Id, ctx.Err())
			return errors.New("searching laptop context cancelled")
		}

		if matchesFilter(filter, laptop) {
			foundLaptop, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			if err := match(foundLaptop); err != nil {
				return err
			}
		}
	}

	return nil
}
