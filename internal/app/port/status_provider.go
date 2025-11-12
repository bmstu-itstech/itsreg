package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type StatusProvider interface {
	Status(ctx context.Context, id bots.BotID) (bots.Status, error)
}
