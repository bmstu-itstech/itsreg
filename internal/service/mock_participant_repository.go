package service

import (
	"context"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockParticipantRepository struct {
	m     map[bots.ParticipantID]bots.Participant
	mutex sync.RWMutex
}

func NewMockParticipantRepository() *MockParticipantRepository {
	return &MockParticipantRepository{
		m: make(map[bots.ParticipantID]bots.Participant),
	}
}

func (r *MockParticipantRepository) UpdateOrCreate(
	ctx context.Context,
	id bots.ParticipantID,
	updateFn func(context.Context, *bots.Participant) error,
) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
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

func (r *MockParticipantRepository) BotThreads(_ context.Context, botID bots.BotID) ([]bots.BotThread, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	threads := make([]bots.BotThread, 0)
	for _, prt := range r.m {
		if prt.ID().BotID() == botID {
			for _, thread := range prt.Threads() {
				uth := bots.NewUserThread(thread, prt.ID().UserID())
				threads = append(threads, uth)
			}
		}
	}
	return threads, nil
}
