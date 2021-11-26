package service

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

// ImageStore is an interface for storing laptop images.
type ImageStore interface {
	// Save stores the image in the store.
	Save(laptopID, extension string, imageData bytes.Buffer) (string, error)
}

type (
	// ImageInfo stores information about an image.
	ImageInfo struct {
		LaptopID  string
		Extension string
		Path      string
	}

	// DiskImageStore stores images on disk and images info on memory.
	DiskImageStore struct {
		mutex        sync.RWMutex
		imagesFolder string
		images       map[string]*ImageInfo
	}
)

// NewDiskImageStore creates a new DiskImageStore.
func NewDiskImageStore(imagesFolder string) *DiskImageStore {
	return &DiskImageStore{
		imagesFolder: imagesFolder,
		images:       make(map[string]*ImageInfo),
	}
}

// Save stores the image in the store.
func (store *DiskImageStore) Save(laptopID, extension string, imageData bytes.Buffer) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error generating image ID: %w", err)
	}

	imagePath := fmt.Sprintf("%s/%s%s", store.imagesFolder, imageID, extension)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("error creating laptop %s image - %s file: %v", laptopID, imageID, err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("error writing laptop %s image - %s to file: %v", laptopID, imageID, err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &ImageInfo{
		LaptopID:  laptopID,
		Extension: extension,
		Path:      imagePath,
	}

	return imageID.String(), nil
}
