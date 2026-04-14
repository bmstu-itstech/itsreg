package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) ScriptMeta(ctx context.Context, id bots.ScriptID) (dto.ScriptMeta, error) {
	row, err := r.getScriptMetaRow(ctx, r.db, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return dto.ScriptMeta{}, port.ErrScriptNotFound
	}
	if err != nil {
		return dto.ScriptMeta{}, err
	}
	return scriptMetaFromRow(row), nil
}
