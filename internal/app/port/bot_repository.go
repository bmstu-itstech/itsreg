package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var (
	ErrBotNotFound      = errors.New("bot not found")
	ErrBotAlreadyExists = errors.New("bot already exists")
)

type BotRepository interface {
	Bot(ctx context.Context, id bots.BotID) (*bots.Bot, error)
	BotsByOwnerID(ctx context.Context, ownerID bots.UserID) ([]*bots.Bot, error)

	SaveBot(ctx context.Context, bot *bots.Bot) error
	UpdateBot(ctx context.Context, bot *bots.Bot) error
}
