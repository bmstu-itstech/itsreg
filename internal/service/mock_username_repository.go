package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockUsernameRepository struct {
	sync.RWMutex
	m map[bots.UserId]bots.Username
}

func NewMockUsernameRepository() *MockUsernameRepository {
	return &MockUsernameRepository{
		m: make(map[bots.UserId]bots.Username),
	}
}

func (r *MockUsernameRepository) Upsert(_ context.Context, id bots.UserId, username bots.Username) error {
	r.Lock()
	defer r.Unlock()
	r.m[id] = username
	return nil
}

func (r *MockUsernameRepository) Username(_ context.Context, id bots.UserId) (bots.Username, error) {
	r.RLock()
	defer r.RUnlock()
	username, ok := r.m[id]
	if !ok {
		return "", fmt.Errorf("%w: %d", bots.ErrUsernameNotFound, id)
	}
	return username, nil
}
