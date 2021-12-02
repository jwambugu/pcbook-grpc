package service

import "sync"

// UserStore is an interface for managing users.
type UserStore interface {
	// Save saves a new user in the store.
	Save(user *User) error
	// FindByUsername fetches a user by their username.
	FindByUsername(username string) (*User, error)
}

// InMemoryUserStore is an in-memory implementation of UserStore.
type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

// NewInMemoryUserStore creates a new InMemoryUserStore.
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// Save saves a new user in the store.
func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.users[user.Username]; exists {
		return ErrRecordExists
	}

	store.users[user.Username] = user.Clone()
	return nil
}

// FindByUsername fetches a user by their username.
func (store *InMemoryUserStore) FindByUsername(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
