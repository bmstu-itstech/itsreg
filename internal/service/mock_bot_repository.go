package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MockBotRepository struct {
	sync.RWMutex
	m map[bots.BotId]bots.Bot
}

func NewMockBotRepository() *MockBotRepository {
	return &MockBotRepository{
		m: make(map[bots.BotId]bots.Bot),
	}
}

func (r *MockBotRepository) Upsert(_ context.Context, bot bots.Bot) error {
	r.Lock()
	defer r.Unlock()
	r.m[bot.Id()] = bot
	return nil
}

func (r *MockBotRepository) Bot(_ context.Context, id bots.BotId) (bots.Bot, error) {
	r.RLock()
	defer r.RUnlock()
	bot, ok := r.m[id]
	if !ok {
		return bots.Bot{}, fmt.Errorf("%w: %s", bots.ErrBotNotFound, id)
	}
	return bot, nil
}

func (r *MockBotRepository) UserBots(_ context.Context, userId bots.UserId) ([]bots.Bot, error) {
	r.RLock()
	defer r.RUnlock()
	res := make([]bots.Bot, 0)
	for _, bot := range r.m {
		if bot.Author() == userId {
			res = append(res, bot)
		}
	}
	return res, nil
}
