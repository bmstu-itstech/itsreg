package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

var ErrThreadNotFound = errors.New("thread not found")

func (r *Repository) UpdateOrCreateParticipant(
	ctx context.Context,
	id bots.ParticipantID,
	updateFn func(context.Context, *bots.Participant) error,
) error {
	const op = "PostgresRepository.UpdateOrCreateParticipant"
	l := r.l.With(
		slog.String("op", op),
		slog.String("bot_id", string(id.BotID())),
		slog.Int64("user_id", int64(id.UserID())),
	)

	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		prt, found, err := r.findParticipant(ctx, tx, id)
		if err != nil {
			return err
		}
		if !found {
			l.InfoContext(ctx, "participant not found, creating a new one")
			prt, err = bots.NewParticipant(id)
			if err != nil {
				return err
			}
		} else {
			l.DebugContext(ctx, "participant found")
		}

		err = updateFn(ctx, prt)
		if err != nil {
			return err
		}

		return r.upsertParticipant(ctx, tx, prt)
	})
}

func (r *Repository) BotThreads(ctx context.Context, botID bots.BotID) ([]bots.BotThread, error) {
	var res []bots.BotThread
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		res, err = r.selectBotThreads(ctx, tx, botID)
		return err
	})
	return res, err
}

//
//
// ОПЕРАЦИИ НАД СУЩНОСТЯМИ ВНУТРИ АГГРЕГАТА
//
//

func (r *Repository) findParticipant(
	ctx context.Context,
	qc sqlx.QueryerContext,
	id bots.ParticipantID,
) (*bots.Participant, bool, error) {
	botID := string(id.BotID())
	userID := int64(id.UserID())

	row, err := r.getParticipantRow(ctx, qc, botID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var thread *bots.Thread
	if row.ActiveThread != nil {
		tID := *row.ActiveThread
		thread, err = r.getThread(ctx, qc, bots.ThreadID(tID))
		if err != nil {
			return nil, true, err
		}
	}

	prt, err := bots.UnmarshallParticipant(row.BotID, row.UserID, thread)
	if err != nil {
		return nil, false, err
	}

	return prt, true, nil
}

func (r *Repository) selectBotThreads(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
) ([]bots.BotThread, error) {
	rows, err := r.selectBotThreadsRows(ctx, qc, string(botID))
	if err != nil {
		return nil, err
	}
	res := make([]bots.BotThread, len(rows))
	for i, row := range rows {
		answers, err2 := r.selectAnswers(ctx, qc, bots.ThreadID(row.ID))
		if err2 != nil {
			return nil, err2
		}
		thread, err2 := bots.UnmarshallThread(row.ID, row.Key, row.State, answers, row.StartedAt)
		if err2 != nil {
			return nil, err2
		}
		res[i] = bots.NewUserThread(thread, bots.UserID(row.UserID))
	}
	return res, nil
}

func (r *Repository) getThread(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID bots.ThreadID,
) (*bots.Thread, error) {
	row, err := r.getThreadRow(ctx, qc, string(threadID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s", ErrThreadNotFound, string(threadID))
	}
	answers, err := r.selectAnswers(ctx, qc, bots.ThreadID(row.ID))
	if err != nil {
		return nil, err
	}
	return bots.UnmarshallThread(row.ID, row.Key, row.State, answers, row.StartedAt)
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

func (r *Repository) upsertParticipant(
	ctx context.Context,
	ec sqlx.ExtContext,
	prt *bots.Participant,
) error {
	botID := prt.ID().BotID()
	userID := prt.ID().UserID()
	prtRow := participantToRow(prt)

	if err := r.upsertParticipantRow(ctx, ec, prtRow); err != nil {
		return err
	}

	thread := prt.ActiveThread()
	if thread != nil {
		thrRow := threadToRow(botID, userID, thread)
		if err := r.upsertThreadRow(ctx, ec, thrRow); err != nil {
			return err
		}

		answerRows := answersToRows(thread.ID(), thread.Answers())
		if err := r.syncAnswerRows(ctx, ec, thread.ID(), answerRows); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) syncAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	threadID bots.ThreadID,
	rows []answerRow,
) error {
	const op = "PostgresRepository.syncAnswerRows"
	l := r.l.With(
		slog.String("op", op),
		slog.String("thread_id", string(threadID)),
	)
	l.DebugContext(ctx, "syncing answer rows")

	dbRows, err := r.selectAnswerRows(ctx, ec, string(threadID))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, answerIdentity, diffcalc.Equal)
	l.DebugContext(ctx, "calculated answer changes",
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
