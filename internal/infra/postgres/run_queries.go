package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getRunRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	runID string,
) (runRow, error) {
	var row runRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, bot_id, token, status, error_msg, started_at, stopped_at
		FROM runs
		WHERE id = $1
		`, runID,
	)
	return row, err
}

func (r *Repository) selectRunsByBotIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]runRow, error) {
	var rows []runRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id, bot_id, token, status, error_msg, started_at, stopped_at
		FROM runs
		WHERE bot_id = $1
		`, botID,
	)
	return rows, err
}

func (r *Repository) selectRunsByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
	filter selectRunsFilter,
) ([]runRow, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT
			r.id, r.bot_id, r.token, r.status, r.error_msg, r.started_at, r.stopped_at
		FROM runs r
		JOIN bots b
			ON b.id = r.bot_id
		WHERE
			b.owner_id = $1
			AND b.deleted_at IS NULL
		`,
	)

	args := []any{ownerID}

	if filter.BotID != nil {
		args = append(args, *filter.BotID)
		queryBuilder.WriteString(fmt.Sprintf(`AND r.bot_id = $%d `, len(args)))
	}

	if filter.Status != nil {
		args = append(args, *filter.Status)
		queryBuilder.WriteString(fmt.Sprintf(`AND r.status = $%d `, len(args)))
	}

	var rows []runRow
	err := pgutils.Select(ctx, qc, &rows, queryBuilder.String(), args...)
	return rows, err
}

func (r *Repository) selectRunsWithStatusStartingOrActive(
	ctx context.Context,
	qc sqlx.QueryerContext,
) ([]runRow, error) {
	var rows []runRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id, bot_id, token, status, error_msg, started_at, stopped_at
		FROM runs
		WHERE
    		status IN ('starting', 'active')
    	`,
	)
	return rows, err
}

func (r *Repository) insertRunRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row runRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO runs
			(id, bot_id, token, status, error_msg, started_at, stopped_at)
		VALUES
			(:id, :bot_id, :token, :status, :error_msg, :started_at, :stopped_at)
		`, row,
	))
}

func (r *Repository) updateRunRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row runRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE runs
		SET
		    status = :status,
			error_msg = :error_msg,
			started_at = :started_at,
			stopped_at = :stopped_at
		WHERE id = :id
		`, row,
	))
}
