package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockBotRepository struct {
	m     map[bots.BotID]bots.Bot
	mutex sync.RWMutex
}

func NewMockBotRepository() *MockBotRepository {
	return &MockBotRepository{
		m: make(map[bots.BotID]bots.Bot),
	}
}

func (r *MockBotRepository) Upsert(_ context.Context, bot bots.Bot) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.m[bot.ID()] = bot
	return nil
}

func (r *MockBotRepository) Bot(_ context.Context, id bots.BotID) (bots.Bot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	bot, ok := r.m[id]
	if !ok {
		return bots.Bot{}, fmt.Errorf("%w: %s", bots.ErrBotNotFound, id)
	}
	return bot, nil
}

func (r *MockBotRepository) UserBots(_ context.Context, userID bots.UserID) ([]bots.Bot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	res := make([]bots.Bot, 0)
	for _, bot := range r.m {
		if bot.Author() == userID {
			res = append(res, bot)
		}
	}
	return res, nil
}
