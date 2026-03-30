package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var ErrThreadAlreadyExists = errors.New("thread already exists")
var ErrUserHasNotThreads = errors.New("user has not threads")

type ThreadRepository interface {
	SaveThread(ctx context.Context, t *bots.Thread) error
	UpdateThread(ctx context.Context, t *bots.Thread) error
	LastUserThread(ctx context.Context, botID bots.BotID, userID bots.UserID) (*bots.Thread, error)
}
