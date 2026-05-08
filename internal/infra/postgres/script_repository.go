package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

// Script возвращает сценарий по его идентификатору. Если сценарий не найден,
// возвращает ошибку port.ErrScriptNotFound.
//
//nolint:gocognit // Метод состоит из повторяющихся блоков кода.
func (r *Repository) Script(ctx context.Context, id bots.ScriptID) (*bots.Script, error) {
	var script *bots.Script
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rScript, err := r.getScriptRow(ctx, tx, id.String())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return port.ErrScriptNotFound
			}
			return fmt.Errorf("getScriptRow: %w", err)
		}

		rEntries, err := r.selectEntriesByScriptIDRows(ctx, tx, id.String())
		if err != nil {
			return fmt.Errorf("selectEntriesByScriptIDRows: %w", err)
		}

		rNodes, err := r.selectNodesByScriptIDRows(ctx, tx, id.String())
		if err != nil {
			return fmt.Errorf("selectNodesByScriptIDRows: %w", err)
		}

		rMessages, err := r.selectMessagesByScriptIDRows(ctx, tx, id.String())
		if err != nil {
			return fmt.Errorf("selectMessagesByScriptIDRows: %w", err)
		}

		rEdges, err := r.selectEdgesByScriptIDRows(ctx, tx, id.String())
		if err != nil {
			return fmt.Errorf("selectEdgesByScriptIDRows: %w", err)
		}

		rOptions, err := r.selectOptionsByScriptIDRows(ctx, tx, id.String())
		if err != nil {
			return fmt.Errorf("selectOptionsByScriptIDRows: %w", err)
		}

		entries, err := entriesFromRows(rEntries)
		if err != nil {
			return err
		}

		nodes, err := nodesFromRows(rNodes, rEdges, rMessages, rOptions)
		if err != nil {
			return err
		}

		script, err = bots.RestoreScript(
			bots.ScriptID(rScript.ID),
			bots.UserID(rScript.OwnerID),
			rScript.Desc,
			nodes,
			entries,
			rScript.CreatedAt,
			rScript.UpdatedAt,
			rScript.DeletedAt,
		)
		if err != nil {
			return err
		}
		return nil
	})
	return script, err
}

// ScriptsByOwnerID возвращает все сценарии, принадлежащие указанному
// пользователю, не включая удаленные. Сценарии сортируются по полю updated_at в
// порядке убывания (последние обновленные сценарии идут первыми).
//
//nolint:gocognit // Метод состоит из повторяющихся блоков кода.
func (r *Repository) ScriptsByOwnerID(ctx context.Context, ownerID bots.UserID) ([]*bots.Script, error) {
	scripts := make([]*bots.Script, 0)
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rScripts, err := r.selectScriptsByOwnerRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectScriptsByOwnerRows: %w", err)
		}

		rEntries, err := r.selectEntriesByOwnerIDRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectEntriesByOwnerIDRows: %w", err)
		}
		mrEntries := hlpr.GroupBy(rEntries, func(m entryRow) string { return m.ScriptID })

		rNodes, err := r.selectNodesByOwnerIDRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectNodesByOwnerIDRows: %w", err)
		}
		mrNodes := hlpr.GroupBy(rNodes, func(m nodeRow) string { return m.ScriptID })

		rMessages, err := r.selectMessagesByOwnerIDRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectMessagesByOwnerIDRows: %w", err)
		}
		mrMessages := hlpr.GroupBy(rMessages, func(m messageRow) string { return m.ScriptID })

		rEdges, err := r.selectEdgesByOwnerIDRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectEdgesByOwnerIDRows: %w", err)
		}
		mrEdges := hlpr.GroupBy(rEdges, func(m edgeRow) string { return m.ScriptID })

		rOptions, err := r.selectOptionsByOwnerIDRows(ctx, tx, ownerID.Int64())
		if err != nil {
			return fmt.Errorf("selectOptionsByOwnerIDRows: %w", err)
		}
		mrOptions := hlpr.GroupBy(rOptions, func(m optionRow) string { return m.ScriptID })

		for _, rScript := range rScripts {
			nodes, err2 := nodesFromRows(
				mrNodes[rScript.ID], mrEdges[rScript.ID], mrMessages[rScript.ID], mrOptions[rScript.ID],
			)
			if err2 != nil {
				return err2
			}

			entries, err2 := entriesFromRows(mrEntries[rScript.ID])
			if err2 != nil {
				return err2
			}

			script, err2 := bots.RestoreScript(
				bots.ScriptID(rScript.ID),
				bots.UserID(rScript.OwnerID),
				rScript.Desc,
				nodes,
				entries,
				rScript.CreatedAt,
				rScript.UpdatedAt,
				rScript.DeletedAt,
			)
			if err2 != nil {
				return err2
			}
			scripts = append(scripts, script)
		}
		return nil
	})
	return scripts, err
}

// SaveScript сохраняет новый сценарий в базе данных. Если сценарий с таким же
// идентификатором уже существует, возвращает ошибку port.ErrScriptAlreadyExists.
func (r *Repository) SaveScript(ctx context.Context, s *bots.Script) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		scriptID := s.ID().String()
		rScript := scriptToRow(s)
		rEntries := entriesToRows(s.Entries(), scriptID)
		rNodes := nodesToRows(s.Nodes(), scriptID)
		rEdges, rMessages, rOptions := decomposeNodesToRows(s.Nodes(), scriptID)

		if err := r.insertScriptRow(ctx, tx, rScript); err != nil {
			if pgutils.IsUniqueViolationError(err) {
				return port.ErrScriptAlreadyExists
			}
			return fmt.Errorf("insertScriptRow: %w", err)
		}

		if err := r.insertNodesRows(ctx, tx, rNodes); err != nil {
			return fmt.Errorf("insertNodesRows: %w", err)
		}

		if err := r.insertEntriesRows(ctx, tx, rEntries); err != nil {
			return fmt.Errorf("insertEntriesRows: %w", err)
		}

		if err := r.insertEdgesRows(ctx, tx, rEdges); err != nil {
			return fmt.Errorf("insertEdgesRows: %w", err)
		}

		if err := r.insertMessagesRows(ctx, tx, rMessages); err != nil {
			return fmt.Errorf("insertMessagesRows: %w", err)
		}

		if err := r.insertOptionsRows(ctx, tx, rOptions); err != nil {
			return fmt.Errorf("insertOptionsRows: %w", err)
		}

		return nil
	})
}

// UpdateScript обновляет существующий сценарий в базе данных. Если сценарий не
// найден, возвращает ошибку port.ErrScriptNotFound.
func (r *Repository) UpdateScript(ctx context.Context, s *bots.Script) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		scriptID := s.ID().String()
		rScript := scriptToRow(s)
		rEntries := entriesToRows(s.Entries(), scriptID)
		rNodes := nodesToRows(s.Nodes(), scriptID)
		rEdges, rMessages, rOptions := decomposeNodesToRows(s.Nodes(), scriptID)

		if err := r.updateScriptRow(ctx, tx, rScript); err != nil {
			if errors.Is(err, pgutils.ErrNoAffectedRows) {
				return port.ErrScriptNotFound
			}
			return fmt.Errorf("updateScriptRow: %w", err)
		}

		if err := r.syncNodesRows(ctx, tx, scriptID, rNodes); err != nil {
			return fmt.Errorf("syncNodesRows: %w", err)
		}

		if err := r.syncEntriesRows(ctx, tx, scriptID, rEntries); err != nil {
			return fmt.Errorf("syncEntriesRows: %w", err)
		}

		if err := r.syncEdgesRows(ctx, tx, scriptID, rEdges); err != nil {
			return fmt.Errorf("syncEdgesRows: %w", err)
		}

		if err := r.syncMessagesRows(ctx, tx, scriptID, rMessages); err != nil {
			return fmt.Errorf("syncMessagesRows: %w", err)
		}

		if err := r.syncOptionsRows(ctx, tx, scriptID, rOptions); err != nil {
			return fmt.Errorf("syncOptionsRows: %w", err)
		}

		return nil
	})
}
