package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type UserRepository interface {
	UpsertUsername(ctx context.Context, id bots.UserID, username bots.Username) error
}
