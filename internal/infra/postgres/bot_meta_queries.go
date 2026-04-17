package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getBotMetaRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	id string,
) (botMetaRow, error) {
	var row botMetaRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, owner_id, script_id, token, deleted_at IS NOT NULL AS deleted
		FROM bots
		WHERE id = $1
		`, id,
	)
	return row, err
}
