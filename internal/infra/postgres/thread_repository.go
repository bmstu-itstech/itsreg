package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

func (r *Repository) SaveThread(ctx context.Context, t *bots.Thread) error {
	const op = "postgres.Repository.SaveThread"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", t.ID().String()),
	)

	tr := threadToRow(t)
	ars := answersToRows(t.ID(), t.Answers())

	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		err := r.insertThreadRow(ctx, tx, tr)
		if pgutils.IsUniqueViolationError(err) {
			l.ErrorContext(ctx, "unique violation error", slog.String("error", err.Error()))
			return port.ErrThreadAlreadyExists
		}
		if err != nil {
			return fmt.Errorf("failed to insert thread: %w", err)
		}

		err = r.insertAnswerRows(ctx, tx, ars)
		if err != nil {
			return fmt.Errorf("failed to insert answer: %w", err)
		}
		return nil
	})
}

func (r *Repository) UpdateThread(ctx context.Context, t *bots.Thread) error {
	tr := threadToRow(t)
	curAnswers := answersToRows(t.ID(), t.Answers())

	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		err := r.updateThreadRow(ctx, tx, tr)
		if err != nil {
			return fmt.Errorf("failed to update thread: %w", err)
		}

		err = r.syncAnswers(ctx, tx, t.ID().String(), curAnswers)
		if err != nil {
			return fmt.Errorf("failed to sync answers: %w", err)
		}
		return nil
	})
}

func (r *Repository) syncAnswers(
	ctx context.Context,
	ec sqlx.ExtContext,
	threadID string,
	curAnswers []answerRow,
) error {
	const op = "postgres.Repository.syncAnswerRows"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", threadID),
	)
	l.DebugContext(ctx, "syncing answer rows")

	prevAnswers, err := r.selectAnswerRows(ctx, ec, threadID)
	if err != nil {
		return fmt.Errorf("failed to select previous answers: %w", err)
	}

	changes := diffcalc.Changes(prevAnswers, curAnswers, answerIdentity, diffcalc.Equal)
	l.DebugContext(ctx, "calculated answers changes",
		slog.String("added", fmt.Sprintf("%v", changes.Added)),
		slog.String("updated", fmt.Sprintf("%v", changes.Updated)),
		slog.String("deleted", fmt.Sprintf("%v", changes.Deleted)),
	)

	if len(changes.Added) > 0 {
		err = r.insertAnswerRows(ctx, ec, changes.Added)
		if err != nil {
			return err
		}
	}

	for _, row := range changes.Updated {
		err = r.updateAnswerRow(ctx, ec, row)
		if err != nil {
			return err
		}
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteAnswerRows(ctx, ec, changes.Deleted)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) LastUserThread(ctx context.Context, botID bots.BotID, userID bots.UserID) (*bots.Thread, error) {
	const op = "postgres.Repository.LastUserThread"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", botID.String()),
		slog.Int64("user_id", userID.Int64()),
	)
	l.DebugContext(ctx, "fetching last user thread")

	var thread *bots.Thread
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		tRow, err := r.getLastUserThreadRow(ctx, tx, botID.String(), userID.Int64())
		if err != nil {
			return fmt.Errorf("failed to get last user thread row: %w", err)
		}

		answers, err := r.selectAnswers(ctx, tx, bots.ThreadID(tRow.ID))
		if err != nil {
			return fmt.Errorf("failed to select answers: %w", err)
		}

		thread, err = bots.RestoreThread(
			bots.ThreadID(tRow.ID),
			bots.BotID(tRow.BotID),
			bots.UserID(tRow.UserID),
			bots.EntryKey(tRow.Key),
			bots.MustNewState(tRow.State),
			answers,
			tRow.StartedAt,
		)
		return err
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: (%s, %d)", port.ErrUserHasNotThreads, botID, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch last user thread: %w", err)
	}
	return thread, nil
}

func (r *Repository) selectAnswers(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID bots.ThreadID,
) (map[bots.State]bots.Message, error) {
	rows, err := r.selectAnswerRows(ctx, qc, string(threadID))
	if err != nil {
		return nil, err
	}
	res := make(map[bots.State]bots.Message)
	for _, row := range rows {
		msg, err2 := bots.NewMessage(row.Text)
		if err2 != nil {
			return nil, err2
		}
		state, err2 := bots.NewState(row.State)
		if err2 != nil {
			return nil, err2
		}
		res[state] = msg
	}
	return res, nil
}
