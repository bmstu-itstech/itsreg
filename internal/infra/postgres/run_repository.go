package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) Run(ctx context.Context, id bots.RunID) (*bots.Run, error) {
	rRun, err := r.getRunRow(ctx, r.db, id.String())
	if errors.Is(sql.ErrNoRows, err) {
		return nil, port.ErrRunNotFound
	}
	if err != nil {
		return nil, err
	}
	return runFromRow(rRun)
}

func (r *Repository) RunsByBotID(ctx context.Context, botID bots.BotID) ([]*bots.Run, error) {
	rRuns, err := r.selectRunsByBotIDRows(ctx, r.db, botID.String())
	if err != nil {
		return nil, err
	}
	return runsFromRows(rRuns)
}

func (r *Repository) RunsByOwnerID(ctx context.Context, ownerID bots.UserID) ([]*bots.Run, error) {
	rRuns, err := r.selectRunsByOwnerIDRows(ctx, r.db, ownerID.Int64())
	if err != nil {
		return nil, err
	}
	return runsFromRows(rRuns)
}

func (r *Repository) SaveRun(ctx context.Context, run *bots.Run) error {
	rRun := runToRow(run)
	err := r.insertRunRow(ctx, r.db, rRun)
	// Ошибка на частичном индексе, гарантирующий, что есть только один Run
	// со статусом starting или running
	if pgutils.IsUniqueViolationError(err) {
		return port.ErrActiveRunAlreadyExists
	}
	return err
}

func (r *Repository) UpdateRun(ctx context.Context, run *bots.Run) error {
	rRun := runToRow(run)
	err := r.updateRunRow(ctx, r.db, rRun)
	if errors.Is(err, sql.ErrNoRows) {
		return port.ErrRunNotFound
	}
	return err
}
