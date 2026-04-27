package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getOwnedMailingRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	mailingID string,
) (ownedMailingModel, error) {
	var row ownedMailingModel
	err := pgutils.Get(ctx, qc, &row, `
		WITH count_recipients AS (
			SELECT 
				mailing_id AS id, 
				count(*)   AS cnt
			FROM mailing_recipients
			GROUP BY id
		),
		count_results AS (
			SELECT
				mailing_id AS id,
				COUNT(CASE WHEN success THEN 1 END) AS success_count,
				COUNT(CASE WHEN NOT success THEN 1 END) AS fail_count,
				count(*) AS results_count
			FROM mailing_results
			GROUP BY id
		)
		SELECT
			m.id, 
			b.owner_id,
			m.bot_id, 
			m.name,
			m.entry_key,
			m.status,
			coalesce(crs.success_count, 0) 	AS success_count,
			coalesce(crs.fail_count, 0) AS fail_count,
			(
				crc.cnt - coalesce(crs.results_count, 0)
			) AS pending_count,
			coalesce(crc.cnt, 0) AS total_count,
			m.created_at,
			m.started_at,
			m.completed_at
		FROM mailings m
		JOIN bots b
			ON b.id = m.bot_id
		JOIN count_recipients crc
			ON crc.id = m.id
		LEFT JOIN count_results crs	
			ON crs.id = m.id
		WHERE 
			m.id = $1
			AND b.deleted_at IS NULL
		ORDER BY created_at DESC
		`, mailingID,
	)
	return row, err
}
