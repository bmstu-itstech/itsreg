package postgres

import (
	"context"

	"github.com/zhikh23/pgutils"

	"github.com/jmoiron/sqlx"
)

func (r *Repository) getBotRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) (botRow, error) {
	var row botRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, owner_id, script_id, token, "desc", created_at, updated_at, deleted_at
		FROM bots
		WHERE id = $1
		`, botID,
	)
	return row, err
}

func (r *Repository) selectBotsByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]botRow, error) {
	var rows []botRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id, owner_id, script_id, token, "desc", created_at, updated_at, deleted_at
		FROM bots
		WHERE 
			owner_id = $1
			AND deleted_at IS NULL
		ORDER BY updated_at DESC
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertBotRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row botRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO bots
			(id, owner_id, script_id, token, "desc", created_at, updated_at, deleted_at) 
		VALUES
			(:id, :owner_id, :script_id, :token, :desc, :created_at, :updated_at, :deleted_at)
		`, row,
	))
}

func (r *Repository) updateBotRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row botRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE bots
		SET
			script_id = :script_id,
			token 	  = :token,
			"desc"	  = :desc,
			updated_at = :updated_at,
			deleted_at = :deleted_at
		WHERE id = :id
		`, row,
	))
}
