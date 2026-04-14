package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) LastUserThread(ctx context.Context, botID bots.BotID, userID bots.UserID) (*bots.Thread, error) {
	var thread *bots.Thread
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rThread, err := r.getLastUserThreadRow(ctx, tx, botID.String(), userID.Int64())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return port.ErrUserHasNotThreads
			}
			return err
		}

		rAnswers, err := r.selectAnswersRows(ctx, tx, rThread.ID)
		if err != nil {
			return err
		}
		answers, err := answersFromRows(rAnswers)
		if err != nil {
			return err
		}

		state, err := bots.NewState(rThread.State)
		if err != nil {
			return err
		}

		thread, err = bots.RestoreThread(
			bots.ThreadID(rThread.ID),
			bots.BotID(rThread.BotID),
			bots.UserID(rThread.UserID),
			bots.EntryKey(rThread.Key),
			state,
			answers,
			rThread.StartedAt,
		)
		if err != nil {
			return err
		}
		return nil
	})
	return thread, err
}

func (r *Repository) SaveThread(ctx context.Context, t *bots.Thread) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rThread := threadToRow(t)
		rAnswers := answersToRows(t.Answers(), t.ID())

		if err := r.insertThreadRow(ctx, tx, rThread); err != nil {
			if pgutils.IsUniqueViolationError(err) {
				return port.ErrThreadAlreadyExists
			}
			return err
		}

		if len(rAnswers) > 0 {
			if err := r.upsertAnswersRows(ctx, tx, rAnswers); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) UpdateThread(ctx context.Context, t *bots.Thread) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rThread := threadToRow(t)
		rAnswers := answersToRows(t.Answers(), t.ID())

		if err := r.updateThreadRow(ctx, tx, rThread); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return port.ErrThreadNotFound
			}
		}

		if len(rAnswers) > 0 {
			if err := r.upsertAnswersRows(ctx, tx, rAnswers); err != nil {
				return err
			}
		}

		return nil
	})
}
