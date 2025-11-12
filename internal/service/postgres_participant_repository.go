package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

var ErrThreadNotFound = errors.New("thread not found")

type PostgresParticipantRepository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewPostgresParticipantRepository(db *sqlx.DB, log *slog.Logger) *PostgresParticipantRepository {
	return &PostgresParticipantRepository{
		db:  db,
		log: log,
	}
}

func (r *PostgresParticipantRepository) UpdateOrCreate(
	ctx context.Context,
	id bots.ParticipantID,
	updateFn func(context.Context, *bots.Participant) error,
) error {
	const op = "PostgresParticipantRepository.UpdateOrCreate"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) BotThreads(ctx context.Context, botID bots.BotID) ([]bots.BotThread, error) {
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

func (r *PostgresParticipantRepository) findParticipant(
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

func (r *PostgresParticipantRepository) selectBotThreads(
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

func (r *PostgresParticipantRepository) getThread(
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

func (r *PostgresParticipantRepository) selectAnswers(
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

func (r *PostgresParticipantRepository) upsertParticipant(
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

func (r *PostgresParticipantRepository) syncAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	threadID bots.ThreadID,
	rows []answerRow,
) error {
	const op = "PostgresParticipantRepository.syncAnswerRows"
	l := r.log.With(
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

//
//
// ОПЕРАЦИИ НАД СТРОКАМИ ТАБЛИЦ
//
//

func (r *PostgresParticipantRepository) getParticipantRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
	userID int64,
) (participantRow, error) {
	const op = "PostgresParticipantRepository.getParticipantRow"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) upsertParticipantRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row participantRow,
) error {
	const op = "PostgresParticipantRepository.upsertParticipantRow"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) getThreadRow(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) (threadRow, error) {
	const op = "PostgresParticipantRepository.getThreadRow"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) selectBotThreadsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID string,
) ([]threadRow, error) {
	const op = "PostgresParticipantRepository.selectBotThreadsRows"
	l := r.log.With(
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
		`,
		botID,
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to query bot thread rows", slog.String("error", err.Error()))
		return nil, fmt.Errorf("selecting bot thread rows: %w", err)
	}
	return rows, nil
}

func (r *PostgresParticipantRepository) upsertThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	const op = "PostgresParticipantRepository.upsertThreadRow"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) selectAnswerRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadID string,
) ([]answerRow, error) {
	const op = "PostgresParticipantRepository.selectAnswerRows"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) insertAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	const op = "PostgresParticipantRepository.insertAnswerRows"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) updateAnswerRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row answerRow,
) error {
	const op = "PostgresParticipantRepository.updateAnswerRow"
	l := r.log.With(
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

func (r *PostgresParticipantRepository) deleteAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []answerRow,
) error {
	const op = "PostgresParticipantRepository.deleteAnswerRows"
	l := r.log.With(
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

//
//
// МОДЕЛИ БАЗЫ ДАННЫХ
//
//

type participantRow struct {
	// PK(BotID, UserID)
	BotID        string  `db:"bot_id"`
	UserID       int64   `db:"user_id"`
	ActiveThread *string `db:"active_thread"`
}

type threadRow struct {
	// PK(ID)
	ID        string    `db:"id"`
	BotID     string    `db:"bot_id"`
	UserID    int64     `db:"user_id"`
	Key       string    `db:"key"`
	State     int       `db:"state"`
	StartedAt time.Time `db:"started_at"`
}

type answerRow struct {
	// PK(ThreadID)
	ThreadID string `db:"thread_id"`
	State    int    `db:"state"`
	Text     string `db:"text"`
}

func answerIdentity(lhs, rhs answerRow) bool {
	return lhs.ThreadID == rhs.ThreadID && lhs.State == rhs.State
}

func participantToRow(prt *bots.Participant) participantRow {
	var activeThreadID *string
	if thread := prt.ActiveThread(); thread != nil {
		s := string(thread.ID())
		activeThreadID = &s
	}
	return participantRow{
		BotID:        string(prt.ID().BotID()),
		UserID:       int64(prt.ID().UserID()),
		ActiveThread: activeThreadID,
	}
}

func threadToRow(botID bots.BotID, userID bots.UserID, thread *bots.Thread) threadRow {
	return threadRow{
		ID:        string(thread.ID()),
		BotID:     string(botID),
		UserID:    int64(userID),
		Key:       string(thread.Key()),
		State:     thread.State().Int(),
		StartedAt: thread.StartedAt(),
	}
}

func answerToRow(threadID bots.ThreadID, state bots.State, msg bots.Message) answerRow {
	return answerRow{
		ThreadID: string(threadID),
		State:    state.Int(),
		Text:     msg.Text(),
	}
}

func answersToRows(threadID bots.ThreadID, answers map[bots.State]bots.Message) []answerRow {
	res := make([]answerRow, 0, len(answers))
	for state, answer := range answers {
		res = append(res, answerToRow(threadID, state, answer))
	}
	return res
}
