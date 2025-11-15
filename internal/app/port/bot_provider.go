package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var ErrBotNotFound = errors.New("bot not found")

type BotProvider interface {
	// Bot возвращает найденного бота или ошибку ErrBotNotFound.
	Bot(ctx context.Context, id bots.BotID) (*bots.Bot, error)

	// UserBots возвращает возможно пустой список ботов, чьи авторы имеют UserID.
	UserBots(ctx context.Context, userID bots.UserID) ([]*bots.Bot, error)

	// EnabledBots возвращает все боты, для которых enabled=true.
	EnabledBots(ctx context.Context) ([]*bots.Bot, error)
}
