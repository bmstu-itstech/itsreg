package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type ThreadProvider interface {
	// BotThreads возвращает все цепочки ответов (треды) заданному боту.
	BotThreads(ctx context.Context, botID bots.BotID) ([]bots.BotThread, error)
}
