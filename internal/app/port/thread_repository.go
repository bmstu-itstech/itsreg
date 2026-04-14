package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var (
	ErrThreadAlreadyExists = errors.New("thread already exists")
	ErrThreadNotFound      = errors.New("thread does not found")
	ErrUserHasNotThreads   = errors.New("user has not threads")
)

type ThreadRepository interface {
	LastUserThread(ctx context.Context, botID bots.BotID, userID bots.UserID) (*bots.Thread, error)
	SaveThread(ctx context.Context, t *bots.Thread) error
	UpdateThread(ctx context.Context, t *bots.Thread) error
}
