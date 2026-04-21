package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) OwnedRun(ctx context.Context, id bots.RunID) (dto.OwnedRun, error) {
	row, err := r.getOwnedRunRow(ctx, r.db, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return dto.OwnedRun{}, port.ErrRunNotFound
	}
	if err != nil {
		return dto.OwnedRun{}, err
	}
	return ownedRunFromRow(row), nil
}
