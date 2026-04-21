package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getOwnedRunRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	runID string,
) (ownedRunRow, error) {
	var row ownedRunRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			r.id, b.owner_id, r.bot_id, r.status, r.error_msg, r.started_at, r.stopped_at
		FROM runs r
		JOIN bots b
			ON b.id = r.bot_id
		WHERE 
			r.id = $1
			AND b.deleted_at IS NULL	
		`, runID,
	)
	return row, err
}
