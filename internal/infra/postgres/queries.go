package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

// CRUD интерфейс для всех строк всех таблиц, связанных с ботом.

func (r *Repository) getBotRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) (botRow, error) {
	var row botRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id,
			token,
			author,
			enabled,
			created_at
		FROM bots
		WHERE
			id = $1
			AND deleted_at IS NULL
		`,
		botID,
	)
	if err != nil {
		return row, fmt.Errorf("selecting bot row: %w", err)
	}
	return row, nil
}

func (r *Repository) selectBotRowsByAuthor(
	ctx context.Context,
	qc sqlx.QueryerContext,
	author int64,
) ([]botRow, error) {
	var rows []botRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id,
			token,
			author,
			enabled,
			created_at
		FROM bots
		WHERE
			author = $1
			AND deleted_at IS NULL
		`,
		author,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting bot rows by author: %w", err)
	}
	return rows, nil
}

func (r *Repository) selectEnabledBotRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
) ([]botRow, error) {
	var rows []botRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id,
			token,
			author,
			enabled,
			created_at
		FROM bots
		WHERE
			enabled = true
			AND deleted_at IS NULL
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting bot rows by author: %w", err)
	}
	return rows, nil
}

func (r *Repository) upsertBotRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row botRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			bots (
				id, 
				token, 
				author,
				enabled,
				created_at
			)
		VALUES (
		    :id,
			:token,
			:author,
			:enabled,
			:created_at
		)
		ON CONFLICT 
			(id)
		DO UPDATE 
		SET
			token      = :token,
			author     = :author,
			enabled    = :enabled,
			created_at = :created_at
		`,
		row,
	))
	if err != nil {
		return fmt.Errorf("upserting bot row: %w", err)
	}
	return nil
}

func (r *Repository) selectEntryRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]entryRow, error) {
	var rows []entryRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			bot_id,
			key,
			start
		FROM entries
		WHERE
			bot_id = $1
		`,
		botID,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting entry rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertEntryRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			entries (
				bot_id, 
				key, 
				start
			) 
		VALUES (
			:bot_id,
			:key,
			:start
		)
		`,
		rows,
	))
	if err != nil {
		return fmt.Errorf("inserting entry rows: %w", err)
	}
	return nil
}

func (r *Repository) updateEntryRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row entryRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE entries
		SET
			start = :start
		WHERE
			bot_id = :bot_id
			AND key = :key
		`,
		row,
	))
	if err != nil {
		return fmt.Errorf("update entry rows: %w", err)
	}
	return nil
}

func (r *Repository) deleteEntryRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	for _, row := range rows {
		err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
			DELETE FROM entries
			WHERE
				bot_id = :bot_id
				AND key = :key
			`,
			row,
		))
		if err != nil {
			return fmt.Errorf("deleting entry rows: %w", err)
		}
	}
	return nil
}

func (r *Repository) selectNodeRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]nodeRow, error) {
	var rows []nodeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			bot_id,
			state,
			title
		FROM nodes
		WHERE
			bot_id = $1
		`,
		botID,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting node rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertNodeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []nodeRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			nodes (
				bot_id,
				state, 
				title
			) 
		VALUES (
			:bot_id,
			:state,
			:title
		)
		`,
		rows,
	))
	if err != nil {
		return fmt.Errorf("inserting node rows: %w", err)
	}
	return nil
}

func (r *Repository) updateNodeRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row nodeRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE nodes
		SET
			title = :title
		WHERE
			bot_id = :bot_id
			AND state = :state
		`,
		row,
	))
	if err != nil {
		return fmt.Errorf("updating node rows: %w", err)
	}
	return err
}

func (r *Repository) deleteNodeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []nodeRow,
) error {
	for _, row := range rows {
		err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
			DELETE FROM nodes
			WHERE
				bot_id = :bot_id
				AND state = :state
			`,
			row,
		))
		if err != nil {
			return fmt.Errorf("deleting node rows: %w", err)
		}
	}
	return nil
}

func (r *Repository) selectEdgeRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	state int,
) ([]edgeRow, error) {
	var rows []edgeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			bot_id,
			state,
			to_state,
			operation,
			pred_type,
			pred_data
		FROM edges
		WHERE
			bot_id = $1
			AND state = $2
		`,
		botID,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting edge rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertEdgeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []edgeRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			edges (
				bot_id, 
				state, 
				to_state, 
				operation, 
				pred_type, 
				pred_data
			) 
		VALUES (
		    :bot_id,
			:state,
			:to_state,
			:operation,
			:pred_type,
			:pred_data
		)
		`,
		rows,
	))
	if err != nil {
		return fmt.Errorf("inserting edge rows: %w", err)
	}
	return nil
}

func (r *Repository) deleteEdgeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID string,
	state int,
) error {
	err := pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
			DELETE FROM edges
			WHERE
				bot_id = $1
				AND state = $2
			`,
		botID,
		state,
	))
	if err != nil {
		return fmt.Errorf("deleting edge rows: %w", err)
	}
	return nil
}

func (r *Repository) selectMessageRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	state int,
) ([]messageRow, error) {
	var rows []messageRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			bot_id,
			state,
			text
		FROM bot_messages
		WHERE
			bot_id = $1
			AND state = $2
		`,
		botID,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting message rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertMessageRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []messageRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			bot_messages(
			    bot_id, 
				state, 
				text
			)
		VALUES (
			:bot_id,
			:state,
			:text
		)
		`,
		rows,
	))
	if err != nil {
		return fmt.Errorf("inserting message rows: %w", err)
	}
	return nil
}

func (r *Repository) deleteMessageRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID string,
	state int,
) error {
	err := pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM bot_messages
		WHERE
		    bot_id = $1
			AND state = $2
		`,
		botID,
		state,
	))
	if err != nil {
		return fmt.Errorf("deleting message rows: %w", err)
	}
	return nil
}

func (r *Repository) selectOptionRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	state int,
) ([]optionRow, error) {
	var rows []optionRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			bot_id,
			state,
			text
		FROM options
		WHERE
		    bot_id = $1
			AND state = $2
		`,
		botID,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting option rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertOptionRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []optionRow,
) error {
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			options (
				bot_id, 
				state, 
				text
			) 
		VALUES (
			:bot_id,
			:state,
			:text
		)
		`,
		rows,
	))
	if err != nil {
		return fmt.Errorf("inserting option rows: %w", err)
	}
	return nil
}

func (r *Repository) deleteOptionRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID string,
	state int,
) error {
	err := pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM options
		WHERE
		    bot_id = $1
			AND state = $2
		`,
		botID,
		state,
	))
	if err != nil {
		return fmt.Errorf("deleting option rows: %w", err)
	}
	return nil
}

func (r *Repository) getParticipantRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	userID int64,
) (participantRow, error) {
	const op = "PostgresRepository.getParticipantRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", botID),
		slog.Int64("user_id", userID),
	)

	l.DebugContext(ctx, "querying participant row")
	var row participantRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			bot_id,
			user_id,
			active_thread
		FROM participants
		WHERE
			bot_id = $1
			AND user_id = $2
		`,
		botID,
		userID,
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to get participant row", slog.String("error", err.Error()))
		return participantRow{}, fmt.Errorf("getting participant row: %w", err)
	}
	return row, nil
}

func (r *Repository) upsertParticipantRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row participantRow,
) error {
	const op = "PostgresRepository.upsertParticipantRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", row.BotID),
		slog.Int64("user_id", row.UserID),
	)

	l.DebugContext(ctx, "upserting participant row")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			participants (
				bot_id, 
				user_id,
				active_thread
			)
		VALUES (
		    :bot_id,
			:user_id,
			:active_thread
		)
		ON CONFLICT 
			(bot_id, user_id)
		DO UPDATE 
		SET
			active_thread = :active_thread
		`,
		row,
	))
	if err != nil {
		l.ErrorContext(ctx, "failed to upsert participant row", slog.String("error", err.Error()))
		return fmt.Errorf("upserting participant row: %w", err)
	}
	return nil
}

func (r *Repository) getThreadRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) (threadRow, error) {
	const op = "PostgresRepository.getThreadRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", threadID),
	)

	l.DebugContext(ctx, "querying thread row")
	var row threadRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id,
			bot_id,
			user_id,
			key,
			state,
			started_at
		FROM threads
		WHERE
		    id = $1
		`,
		threadID,
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to query thread row", slog.String("error", err.Error()))
		return threadRow{}, fmt.Errorf("querying thread row: %w", err)
	}
	return row, nil
}

func (r *Repository) selectBotThreadsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]threadRow, error) {
	const op = "PostgresRepository.selectBotThreadsRows"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", botID),
	)

	l.DebugContext(ctx, "querying bot thread rows")
	var rows []threadRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id,
			bot_id,
			user_id,
			key,
			state,
			started_at
		FROM threads
		WHERE
			bot_id = $1
		ORDER BY started_at DESC
		`,
		botID,
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to query bot thread rows", slog.String("error", err.Error()))
		return nil, fmt.Errorf("selecting bot thread rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) upsertThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	const op = "PostgresRepository.upsertThreadRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("id", row.ID),
	)

	l.DebugContext(ctx, "upserting thread row")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO 
			threads (
				id, 
				bot_id, 
				user_id, 
				key, 
				state, 
				started_at
			)	 
		VALUES (
			:id,
			:bot_id,
			:user_id,
			:key,
			:state,
			:started_at
		)
		ON CONFLICT (id)
		DO UPDATE SET
			state = :state
		`,
		row,
	))
	if err != nil {
		l.ErrorContext(ctx, "failed to upsert thread row", slog.String("error", err.Error()))
		return fmt.Errorf("upserting thread row: %w", err)
	}
	return nil
}

func (r *Repository) selectAnswerRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) ([]answerRow, error) {
	const op = "PostgresRepository.selectAnswerRows"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", threadID),
	)

	l.DebugContext(ctx, "querying answer rows")
	var rows []answerRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			thread_id,
			state,
			text
		FROM answers
		WHERE
			thread_id = $1
		`,
		threadID,
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to query answer rows", slog.String("error", err.Error()))
		return nil, fmt.Errorf("selecting answer rows: %w", err)
	}
	return rows, nil
}

func (r *Repository) insertAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	const op = "PostgresRepository.insertAnswerRows"
	l := r.l.With(
		slog.String("op", op),
		slog.Int("rows", len(rows)),
	)

	l.DebugContext(ctx, "inserting answer rows")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			answers (
				thread_id, 
			    state,
				text
			)
		VALUES (
			:thread_id,
			:state,
			:text
		)
		`,
		rows,
	))
	if err != nil {
		l.ErrorContext(ctx, "failed to insert answer rows", slog.String("error", err.Error()))
		return fmt.Errorf("inserting answer rows: %w", err)
	}
	return nil
}

func (r *Repository) updateAnswerRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row answerRow,
) error {
	const op = "PostgresRepository.updateAnswerRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", row.ThreadID),
		slog.Int("state", row.State),
	)

	l.DebugContext(ctx, "updating answer row")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE answers
		SET
			text = :text
		WHERE
			thread_id = :thread_id
			AND state = :state
		`,
		row,
	))
	if err != nil {
		l.ErrorContext(ctx, "failed to update answer row", slog.String("error", err.Error()))
		return fmt.Errorf("updating answer rows: %w", err)
	}
	return nil
}

func (r *Repository) deleteAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	const op = "PostgresRepository.deleteAnswerRows"
	l := r.l.With(
		slog.String("op", op),
		slog.Int("rows", len(rows)),
	)

	l.DebugContext(ctx, "deleting answer rows")
	for _, row := range rows {
		err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
			DELETE FROM answers
			WHERE
				thread_id = :thread_id
				AND state = :state
			`,
			row,
		))
		if err != nil {
			l.ErrorContext(ctx, "failed to delete answer rows", slog.String("error", err.Error()))
			return fmt.Errorf("deleting answer rows: %w", err)
		}
	}
	return nil
}

func (r *Repository) softDeleteBotRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID string,
) error {
	const op = "PostgresRepository.softDeleteBotRow"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", botID),
	)

	l.DebugContext(ctx, "deleting bot row")
	err := pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE bots
		SET deleted_at = now()
		WHERE id = $1
		`,
		botID,
	))
	if err != nil {
		l.ErrorContext(ctx, "failed to soft delete bot", slog.String("error", err.Error()))
		return fmt.Errorf("deleting bot row: %w", err)
	}
	return nil
}
