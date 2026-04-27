package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getMailingRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	mailingID string,
) (mailingRow, error) {
	var row mailingRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, bot_id, name, entry_key, status, created_at, started_at, completed_at
		FROM mailings
		WHERE id = $1
		`, mailingID,
	)
	return row, err
}

func (r *Repository) selectMailingsByBotIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]mailingRow, error) {
	var rows []mailingRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id, bot_id, name, entry_key, status, created_at, started_at, completed_at
		FROM mailings
		WHERE bot_id = $1
		ORDER BY created_at DESC
		`, botID,
	)
	return rows, err
}

func (r *Repository) selectMailingsByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
	filter selectMailingsFilter,
) ([]mailingRow, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT
			m.id, m.bot_id, m.name, m.entry_key, m.status, m.created_at, m.started_at, m.completed_at
		FROM mailings m
		JOIN bots b
			ON b.id = m.bot_id
		WHERE
			b.owner_id = $1
			AND b.deleted_at IS NULL
		ORDER BY created_at DESC
		`,
	)

	args := []any{ownerID}

	if filter.BotID != nil {
		args = append(args, *filter.BotID)
		queryBuilder.WriteString(fmt.Sprintf(` AND m.bot_id = $%d`, len(args)))
	}

	if filter.Status != nil {
		args = append(args, *filter.Status)
		queryBuilder.WriteString(fmt.Sprintf(` AND m.status = $%d`, len(args)))
	}

	var rows []mailingRow
	err := pgutils.Select(ctx, qc, &rows, queryBuilder.String(), args...)
	return rows, err
}

func (r *Repository) insertMailingRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row mailingRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO mailings
			(id, bot_id, name, entry_key, status, created_at, started_at, completed_at) 
		VALUES
			(:id, :bot_id, :name, :entry_key, :status, :created_at, :started_at, :completed_at)
		`, row,
	))
}

func (r *Repository) updateMailingRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row mailingRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE mailings
		SET
		    status = :status,
			started_at = :started_at,
			completed_at = :completed_at
		WHERE id = :id
		`, row,
	))
}

func (r *Repository) selectMailingRecipientsByMailingIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	mailingID string,
) ([]mailingRecipientRow, error) {
	var rows []mailingRecipientRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			mailing_id, user_id
		FROM mailing_recipients
		WHERE mailing_id = $1
		`, mailingID,
	)
	return rows, err
}

func (r *Repository) selectMailingRecipientsByMailingIDsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ids []string,
) ([]mailingRecipientRow, error) {
	var rows []mailingRecipientRow
	if len(ids) == 0 {
		return rows, nil
	}
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			mailing_id, user_id
		FROM mailing_recipients
		WHERE mailing_id = ANY($1)
		`, pq.Array(ids),
	)
	return rows, err
}

func (r *Repository) upsertMailingRecipientsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []mailingRecipientRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := pgutils.NamedExec(ctx, ec, `
		INSERT INTO mailing_recipients
			(mailing_id, user_id)
		VALUES
			(:mailing_id, :user_id)
		ON CONFLICT (mailing_id, user_id) DO NOTHING
		`, rows,
	)
	return err
}

func (r *Repository) selectMailingResultsByMailingIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	mailingID string,
) ([]mailingResultRow, error) {
	var rows []mailingResultRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			mailing_id, user_id, success, error_msg
		FROM mailing_results
		WHERE mailing_id = $1
		`, mailingID,
	)
	return rows, err
}

func (r *Repository) selectMailingResultsByMailingIDsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ids []string,
) ([]mailingResultRow, error) {
	var rows []mailingResultRow
	if len(ids) == 0 {
		return rows, nil
	}
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			mailing_id, user_id, success, error_msg
		FROM mailing_results
		WHERE mailing_id = ANY($1)
		`, pq.Array(ids),
	)
	return rows, err
}

func (r *Repository) upsertMailingResultsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []mailingResultRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO mailing_results
			(mailing_id, user_id, success, error_msg)
		VALUES
			(:mailing_id, :user_id, :success, :error_msg)
		ON CONFLICT (mailing_id, user_id) DO UPDATE
			SET success = EXCLUDED.success, error_msg = EXCLUDED.error_msg
		`, rows,
	))
}
