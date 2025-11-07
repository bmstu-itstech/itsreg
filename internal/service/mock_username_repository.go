package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockUsernameRepository struct {
	m     map[bots.UserID]bots.Username
	mutex sync.RWMutex
}

func NewMockUsernameRepository() *MockUsernameRepository {
	return &MockUsernameRepository{
		m: make(map[bots.UserID]bots.Username),
	}
}

func (r *MockUsernameRepository) Upsert(_ context.Context, id bots.UserID, username bots.Username) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.m[id] = username
	return nil
}

func (r *MockUsernameRepository) Username(_ context.Context, id bots.UserID) (bots.Username, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	username, ok := r.m[id]
	if !ok {
		return "", fmt.Errorf("%w: %d", bots.ErrUsernameNotFound, id)
	}
	return username, nil
}
