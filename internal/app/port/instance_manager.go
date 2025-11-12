package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var ErrRunningInstanceNotFound = errors.New("running instance not found")

type InstanceManager interface {
	Start(ctx context.Context, id bots.BotID, token bots.Token) error
	Stop(ctx context.Context, id bots.BotID) error
}
