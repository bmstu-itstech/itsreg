package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) selectThreadTableRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]threadTableRow, error) {
	var rows []threadTableRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
  			t.id, 
  			t.key,
			t.user_id, 
			coalesce(u.username, '') AS username,
			t.started_at AS ts,
			n.title AS Header,
  			a.text AS Value
		FROM threads t
		LEFT JOIN users u
			ON u.id = t.user_id
		JOIN nodes n
			ON n.bot_id = t.bot_id
		LEFT JOIN answers a 
			ON a.thread_id = t.id
			AND a.state = n.state
		WHERE t.bot_id = $1
		ORDER BY t.id, n.state;
		`, botID,
	)
	return rows, err
}
