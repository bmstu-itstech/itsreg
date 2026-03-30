package postgres

import (
	"context"
	"fmt"

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
		    :ID,
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
			:Key,
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
			AND Key = :Key
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
				AND Key = :Key
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

func (r *Repository) getThreadRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) (threadRow, error) {
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
	return row, err
}

func (r *Repository) getLastUserThreadRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	userID int64,
) (threadRow, error) {
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
			bot_id = $1
			AND user_id = $2
		ORDER BY started_at DESC
		LIMIT 1
		`,
		botID,
		userID,
	)
	return row, err
}

func (r *Repository) selectBotThreadsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]threadRow, error) {
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
	return rows, err
}

func (r *Repository) insertThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO 
			threads (
				id, 
				bot_id, 
				user_id, 
				Key, 
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
		`,
		row,
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
			state = :state
		WHERE id = :id
		`,
		row,
	))
}

func (r *Repository) selectAnswerRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) ([]answerRow, error) {
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
	return rows, err
}

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
			u.username, 
			t.started_at AS ts,
			n.title AS Header,
  			a.text AS Value
		FROM threads t
		JOIN users u
			ON u.id = t.user_id
		JOIN nodes n
			ON n.bot_id = t.bot_id
		LEFT JOIN answers a 
			ON a.thread_id = t.id
			AND a.state = n.state
		WHERE t.bot_id = $1
		ORDER BY t.id, n.state;
		`,
		botID,
	)
	return rows, err
}

func (r *Repository) insertAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
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
}

func (r *Repository) updateAnswerRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row answerRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE answers
		SET
			text = :text
		WHERE
			thread_id = :thread_id
			AND state = :state
		`,
		row,
	))
}

func (r *Repository) deleteAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
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
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE bots
		SET deleted_at = now()
		WHERE id = $1
		`,
		botID,
	))
}

func (r *Repository) updateUsername(
	ctx context.Context,
	ec sqlx.ExtContext,
	userID int64,
	username string,
) error {
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		INSERT INTO 
			users (
				id, 
				username, 
				updated_at
			) 
		VALUES 
			($1, $2, now())	
		ON CONFLICT (id)
			DO UPDATE SET 
				username = $2, 
				updated_at = now()
		`,
		userID,
		username,
	))
}
