package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) getScriptRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) (scriptRow, error) {
	var row scriptRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			id, owner_id, "desc", created_at, updated_at, deleted_at
		FROM scripts
		WHERE id = $1
		`, scriptID,
	)
	return row, err
}

func (r *Repository) selectScriptsByOwnerRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]scriptRow, error) {
	var rows []scriptRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			id, owner_id, "desc", created_at, updated_at, deleted_at
		FROM scripts
		WHERE 
			owner_id = $1
			AND deleted_at IS NULL
		ORDER BY updated_at DESC
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertScriptRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row scriptRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO scripts
			(id, owner_id, "desc", created_at, updated_at, deleted_at) 
		VALUES (:id, :owner_id, :desc, :created_at, :updated_at, :deleted_at)
		`, row,
	))
}

func (r *Repository) updateScriptRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row scriptRow,
) error {
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE scripts
		SET
			"desc" = :desc,
			updated_at = :updated_at,
			deleted_at = :deleted_at
		WHERE id = :id
		`, row,
	))
}

func (r *Repository) selectEntriesByScriptIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) ([]entryRow, error) {
	var rows []entryRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			script_id, key, start
		FROM entries
		WHERE script_id = $1
		`, scriptID,
	)
	return rows, err
}

func (r *Repository) selectEntriesByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]entryRow, error) {
	var rows []entryRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			e.script_id, e.key, e.start
		FROM entries e
		JOIN scripts s
			ON s.id = e.script_id
		WHERE s.owner_id = $1
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertEntriesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO entries
			(script_id, key, start) 
		VALUES 
			(:script_id, :key, :start)
		`, rows,
	))
}

func (r *Repository) updateEntriesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := entriesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE entries e
		SET start = v.start
		FROM unnest($1::text[], $2::text[], $3::int[]) 
			AS v(script_id, key, start)
		WHERE
			e.script_id = v.script_id
			AND e.key = v.key
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.Keys), pq.Array(soa.Starts),
	))
}

func (r *Repository) deleteEntriesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := entriesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM entries e
		USING unnest($1::text[], $2::text[]) 
			AS v(script_id, key)
		WHERE 
			e.script_id = v.script_id
			AND e.key = v.key
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.Keys),
	))
}

func (r *Repository) selectNodesByScriptIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) ([]nodeRow, error) {
	var rows []nodeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			script_id, state, title
		FROM nodes
		WHERE script_id = $1
		`, scriptID,
	)
	return rows, err
}

func (r *Repository) selectNodesByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]nodeRow, error) {
	var rows []nodeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			n.script_id, n.state, n.title
		FROM nodes n
		JOIN scripts s
			ON s.id = n.script_id
		WHERE owner_id = $1
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertNodesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []nodeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO nodes
			(script_id, state, title) 
		VALUES 
			(:script_id, :state, :title)
		`, rows,
	))
}

func (r *Repository) updateNodesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []nodeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := nodesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE nodes n
		SET title = v.title
		FROM unnest($1::text[], $2::int[], $3::text[]) 
			AS v(script_id, state, title)
		WHERE
			n.script_id = v.script_id
			AND n.state = v.state
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Titles),
	))
}

func (r *Repository) deleteNodesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []nodeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := nodesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM nodes n
		USING unnest($1::text[], $2::int[]) 
			AS v(script_id, state)
		WHERE 
			n.script_id = v.script_id
			AND n.state = v.state
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States),
	))
}

func (r *Repository) selectEdgesByScriptIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) ([]edgeRow, error) {
	var rows []edgeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			e.script_id, e.state, e.index, e.to_state, e.operation, e.pred_type, e.pred_data
		FROM edges e
		JOIN nodes n
			ON n.script_id = e.script_id
			AND n.state = e.state
		WHERE e.script_id = $1
		ORDER BY e.state, e.index
		`, scriptID,
	)
	return rows, err
}

func (r *Repository) selectEdgesByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]edgeRow, error) {
	var rows []edgeRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			e.script_id, e.state, e.index, e.to_state, e.operation, e.pred_type, e.pred_data
		FROM edges e
		JOIN scripts s
			ON s.id = e.script_id
		WHERE s.owner_id = $1
		ORDER BY e.script_id, e.state, e.index
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertEdgesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []edgeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO edges
			(script_id, state, index, to_state, operation, pred_type, pred_data) 
		VALUES 
			(:script_id, :state, :index, :to_state, :operation, :pred_type, :pred_data)
		`, rows,
	))
}

func (r *Repository) updateEdgesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []edgeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := edgesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE edges e
		SET 
		    to_state = v.to_state,
		    operation = v.operation,
			pred_type = v.pred_type,
			pred_data = v.pred_data
		FROM unnest(
			$1::text[],
			$2::int[],
			$3::int[],
			$4::int[],
			$5::edge_operation[],
			$6::edge_pred_type[],
			$7::text[]
		) AS v(script_id, state, index, to_state, operation, pred_type, pred_data)
		WHERE 
			e.script_id = v.script_id
			AND e.state = v.state
			AND e.index = v.index
		`,
		pq.Array(soa.ScriptIDs),
		pq.Array(soa.States),
		pq.Array(soa.Indices),
		pq.Array(soa.ToStates),
		pq.Array(soa.Operations),
		pq.Array(soa.PredTypes),
		pq.Array(soa.PredData),
	))
}

func (r *Repository) deleteEdgesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []edgeRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := edgesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM edges e
		USING unnest($1::text[], $2::int[], $3::int[]) 
			AS v(script_id, state, idx)
		WHERE
			e.script_id = v.script_id
			AND e.state = v.state
			AND e.index = v.idx
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Indices),
	))
}

func (r *Repository) selectMessagesByScriptIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) ([]messageRow, error) {
	var rows []messageRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			m.script_id, m.state, m.index, m.text
		FROM messages m
		JOIN nodes n
			ON n.script_id = m.script_id
			AND n.state = m.state
		WHERE m.script_id = $1
		ORDER BY m.state, m.index
		`, scriptID,
	)
	return rows, err
}

func (r *Repository) selectMessagesByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]messageRow, error) {
	var rows []messageRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			m.script_id, m.state, m.index, m.text
		FROM messages m
		JOIN scripts s
			ON s.id = m.script_id
		WHERE s.owner_id = $1
		ORDER BY m.script_id, m.state, m.index
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertMessagesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []messageRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO messages
			(script_id, state, index, text)
		VALUES 
			(:script_id, :state, :index, :text)
		`, rows,
	))
}

func (r *Repository) updateMessagesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []messageRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := messagesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE messages m
		SET text = v.text
		FROM unnest($1::text[], $2::int[], $3::int[], $4::text[]) 
			AS v(script_id, state, index, text)
		WHERE 
			m.script_id = v.script_id
			AND m.state = v.state
			AND m.index = v.index
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Indices), pq.Array(soa.Texts),
	))
}

func (r *Repository) deleteMessagesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []messageRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := messagesRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM messages m
		USING unnest($1::text[], $2::int[], $3::int[]) 
			AS v(script_id, state, idx)
		WHERE
			m.script_id = v.script_id
			AND m.state = v.state
			AND m.index = v.idx
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Indices),
	))
}

func (r *Repository) selectOptionsByScriptIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	scriptID string,
) ([]optionRow, error) {
	var rows []optionRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			o.script_id, o.state, o.index, o.text
		FROM options o
		JOIN nodes n
			ON n.script_id = o.script_id
			AND n.state = o.state
		WHERE o.script_id = $1
		ORDER BY o.state, o.index
		`, scriptID,
	)
	return rows, err
}

func (r *Repository) selectOptionsByOwnerIDRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	ownerID int64,
) ([]optionRow, error) {
	var rows []optionRow
	err := pgutils.Select(ctx, qc, &rows, `
		SELECT
			o.script_id, o.state, o.index, o.text
		FROM options o
		JOIN scripts s
			ON s.id = o.script_id
		WHERE s.owner_id = $1
		ORDER BY o.script_id, o.state, o.index
		`, ownerID,
	)
	return rows, err
}

func (r *Repository) insertOptionsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []optionRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	return pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO options
			(script_id, state, index, text)
		VALUES 
			(:script_id, :state, :index, :text)
		`, rows,
	))
}

func (r *Repository) updateOptionsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []optionRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := optionsRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		UPDATE options o
		SET text = v.text
		FROM unnest($1::text[], $2::int[], $3::int[], $4::text[]) 
			AS v(script_id, state, index, text)
		WHERE 
			o.script_id = v.script_id
			AND o.state = v.state
			AND o.index = v.index
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Indices), pq.Array(soa.Texts),
	))
}

func (r *Repository) deleteOptionsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []optionRow,
) error {
	if len(rows) == 0 {
		return nil
	}
	soa := optionsRowsAoSToSoA(rows)
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM options o
		USING unnest($1::text[], $2::int[], $3::int[]) 
			AS v(script_id, state, idx)
		WHERE
			o.script_id = v.script_id
			AND o.state = v.state
			AND o.index = v.idx
		`, pq.Array(soa.ScriptIDs), pq.Array(soa.States), pq.Array(soa.Indices),
	))
}
