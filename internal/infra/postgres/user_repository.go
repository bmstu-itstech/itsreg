package postgres

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func (r *Repository) UpsertUsername(ctx context.Context, id bots.UserID, username bots.Username) error {
	return r.updateUsername(ctx, r.db, id.Int64(), username.String())
}
