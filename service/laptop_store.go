package service

import (
	"errors"
	"fmt"
	"github.com/jwambugu/pcbook-grpc/protos/pb"
	"sync"

	"github.com/jinzhu/copier"
)

// ErrRecordExists is an error that is returned when a record already exists
var ErrRecordExists = errors.New("record already exists")

// LaptopStore is an interface for storing laptops
type LaptopStore interface {
	// Save saves a laptop in the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by its id
	Find(id string) (*pb.Laptop, error)
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

// Save saves a laptop in the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutext.Lock()
	defer store.mutext.Unlock()

	if _, ok := store.data[laptop.Id]; ok {
		return ErrRecordExists
	}

	newLaptop := &pb.Laptop{}
	if err := copier.Copy(newLaptop, laptop); err != nil {
		return fmt.Errorf("error copying laptop: %v", err)
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

	foundLaptop := &pb.Laptop{}
	if err := copier.Copy(foundLaptop, laptop); err != nil {
		return nil, fmt.Errorf("error copying laptop: %v", err)
	}

	return foundLaptop, nil
}
