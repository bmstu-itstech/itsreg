package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type ProcessHandler interface {
	Process(ctx context.Context, botID bots.BotID, userID bots.UserID, msg bots.Message) error
}
