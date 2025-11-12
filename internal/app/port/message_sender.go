package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MessageSender interface {
	Send(ctx context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage) error
}
