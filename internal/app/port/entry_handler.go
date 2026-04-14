package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type EntryHandler interface {
	Entry(ctx context.Context, botID bots.BotID, userID bots.UserID, username bots.Username, key bots.EntryKey) error
}
