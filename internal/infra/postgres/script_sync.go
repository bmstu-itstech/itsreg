package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/bmstu-itstech/itsreg/pkg/diffcalc"
)

func entryIdentity(a, b entryRow) bool {
	return a.ScriptID == b.ScriptID && a.Key == b.Key
}

func nodeIdentity(a, b nodeRow) bool {
	return a.ScriptID == b.ScriptID && a.State == b.State
}

func edgeIdentity(a, b edgeRow) bool {
	return a.ScriptID == b.ScriptID && a.State == b.State && a.Index == b.Index
}

func messageIdentity(a, b messageRow) bool {
	return a.ScriptID == b.ScriptID && a.State == b.State && a.Index == b.Index
}

func optionIdentity(a, b optionRow) bool {
	return a.ScriptID == b.ScriptID && a.State == b.State && a.Index == b.Index
}

// syncEntriesRows синхронизирует таблицу entries для заданного scriptID.
func (r *Repository) syncEntriesRows(ctx context.Context, ec sqlx.ExtContext, scriptID string, rows []entryRow) error {
	dbRows, err := r.selectEntriesByScriptIDRows(ctx, ec, scriptID)
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, entryIdentity, diffcalc.Equal[entryRow])

	if err2 := r.insertEntriesRows(ctx, ec, changes.Added); err2 != nil {
		return err2
	}

	if err2 := r.updateEntriesRows(ctx, ec, changes.Updated); err2 != nil {
		return err2
	}

	if err2 := r.deleteEntriesRows(ctx, ec, changes.Deleted); err2 != nil {
		return err2
	}

	return nil
}

// syncNodesRows синхронизирует таблицу nodes для заданного scriptID.
func (r *Repository) syncNodesRows(ctx context.Context, ec sqlx.ExtContext, scriptID string, rows []nodeRow) error {
	dbRows, err := r.selectNodesByScriptIDRows(ctx, ec, scriptID)
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, nodeIdentity, diffcalc.Equal[nodeRow])

	if err2 := r.insertNodesRows(ctx, ec, changes.Added); err2 != nil {
		return err2
	}

	if err2 := r.updateNodesRows(ctx, ec, changes.Updated); err2 != nil {
		return err2
	}

	if err2 := r.deleteNodesRows(ctx, ec, changes.Deleted); err2 != nil {
		return err2
	}

	return nil
}

// syncEdgesRows синхронизирует таблицу edges для заданного scriptID и state.
func (r *Repository) syncEdgesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	scriptID string,
	rows []edgeRow,
) error {
	dbRows, err := r.selectEdgesByScriptIDRows(ctx, ec, scriptID)
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, edgeIdentity, diffcalc.Equal[edgeRow])

	if err2 := r.insertEdgesRows(ctx, ec, changes.Added); err2 != nil {
		return err2
	}

	if err2 := r.updateEdgesRows(ctx, ec, rows); err2 != nil {
		return err2
	}

	if err2 := r.deleteEdgesRows(ctx, ec, changes.Deleted); err2 != nil {
		return err2
	}

	return nil
}

// syncMessagesRows синхронизирует таблицу messages для заданного scriptID и state.
func (r *Repository) syncMessagesRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	scriptID string,
	rows []messageRow,
) error {
	dbRows, err := r.selectMessagesByScriptIDRows(ctx, ec, scriptID)
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, messageIdentity, diffcalc.Equal[messageRow])

	if err2 := r.insertMessagesRows(ctx, ec, changes.Added); err2 != nil {
		return err2
	}

	if err2 := r.updateMessagesRows(ctx, ec, changes.Updated); err2 != nil {
		return err2
	}

	if err2 := r.deleteMessagesRows(ctx, ec, changes.Deleted); err2 != nil {
		return err2
	}

	return nil
}

// syncOptionsRows синхронизирует таблицу options для заданного scriptID и state.
func (r *Repository) syncOptionsRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	scriptID string,
	rows []optionRow,
) error {
	dbRows, err := r.selectOptionsByScriptIDRows(ctx, ec, scriptID)
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, optionIdentity, diffcalc.Equal[optionRow])

	if err2 := r.insertOptionsRows(ctx, ec, changes.Added); err2 != nil {
		return err2
	}

	if err2 := r.updateOptionsRows(ctx, ec, changes.Updated); err2 != nil {
		return err2
	}

	if err2 := r.deleteOptionsRows(ctx, ec, changes.Deleted); err2 != nil {
		return err2
	}

	return nil
}
