package service

import (
	"context"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockParticipantRepository struct {
	sync.RWMutex
	m map[bots.ParticipantId]bots.Participant
}

func NewMockParticipantRepository() *MockParticipantRepository {
	return &MockParticipantRepository{
		m: make(map[bots.ParticipantId]bots.Participant),
	}
}

func (r *MockParticipantRepository) UpdateOrCreate(
	ctx context.Context,
	id bots.ParticipantId,
	updateFn func(context.Context, *bots.Participant) error,
) error {
	r.Lock()
	defer r.Unlock()
	prt, ok := r.m[id]
	if !ok {
		newPrt, err := bots.NewParticipant(id)
		if err != nil {
			return err
		}
		prt = *newPrt
	}
	err := updateFn(ctx, &prt)
	if err != nil {
		return err
	}
	r.m[id] = prt
	return nil
}
