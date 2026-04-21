package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getScriptMetaRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	id string,
) (scriptMetaRow, error) {
	var row scriptMetaRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, owner_id, "desc", deleted_at IS NOT NULL AS deleted
		FROM scripts
		WHERE id = $1
		`, id,
	)
	return row, err
}
