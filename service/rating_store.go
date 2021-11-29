package service

import "sync"

// RatingStore is an interface for storing and retrieving laptop ratings.
type RatingStore interface {
	// Add adds a new laptop to the store.
	Add(laptopID string, score float64) (*Rating, error)
}

// Rating contains the rating information for a given laptop.
type Rating struct {
	Count uint32
	Sum   float64
}

// InMemoryRatingStore stores laptop ratings in memory.
type InMemoryRatingStore struct {
	mutext  sync.RWMutex
	ratings map[string]*Rating
}

// NewInMemoryRatingStore creates a new InMemoryRatingStore.
func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		ratings: make(map[string]*Rating),
	}
}

// Add adds a new laptop to the store.
func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutext.Lock()
	defer store.mutext.Unlock()

	rating := store.ratings[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.ratings[laptopID] = rating
	return rating, nil
}
