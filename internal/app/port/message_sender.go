package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var ErrUserBlockedBot = errors.New("user blocked bot")

type MessageSender interface {
	Send(ctx context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage) error
}
