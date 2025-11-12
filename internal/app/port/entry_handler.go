package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type EntryHandler interface {
	Entry(ctx context.Context, botID bots.BotID, userID bots.UserID, key bots.EntryKey) error
}
