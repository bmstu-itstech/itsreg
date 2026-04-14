package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getLastUserThreadRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
	userID int64,
) (threadRow, error) {
	var row threadRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, script_id, user_id, key, state, started_at, updated_at
		FROM threads
		`, scriptID, userID,
	)
	return row, err
}

func (r *Repository) insertThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO threads 
			(id, script_id, user_id, key, state, started_at, updated_at)
		VALUES 
			(:id, :script_id, :user_id, :key, :state, :started_at, :updated_at)
		`, row,
	))
}

func (r *Repository) updateThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE threads
		SET
			state = :state,
			updated_at = :updatedAt
		WHERE id = :id
		`, row,
	))
}

func (r *Repository) selectAnswersRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) ([]answerRow, error) {
	var rows []answerRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			thread_id, state, text
		FROM answers
		WHERE thread_id = $1
		`, threadID,
	)
	return rows, err
}

func (r *Repository) upsertAnswersRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO answers
			(thread_id, state, text)
		VALUES
			(:thread_id, :state, :text)
		ON CONFLICT DO UPDATE 
		SET
			text = :text
		`, rows,
	))
}

func (r *Repository) deleteAnswersRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	threadID string,
	states []int,
) error {
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM answers
		WHERE
			thread_id = $1
			AND state = ANY($2)
		`, threadID, states,
	))
}
