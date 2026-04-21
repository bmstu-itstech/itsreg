package port

import (
	"context"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type RateLimiter interface {
	Wait(ctx context.Context, token bots.Token, now time.Time) (time.Duration, error)
}
