package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

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
	id bots.ParticipantId,
	updateFn func(context.Context, *bots.Participant) error,
) error {
	const op = "PostgresParticipantRepository.UpdateOrCreate"
	l := r.log.With(
		slog.String("op", op),
		slog.String("bot_id", string(id.BotId())),
		slog.Int64("user_id", int64(id.UserId())),
	)

	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		prt, found, err := r.findParticipant(ctx, tx, id)
		if err != nil {
			return err
		}
		if !found {
			l.Info("participant not found, creating a new one")
			prt, err = bots.NewParticipant(id)
			if err != nil {
				return err
			}
		} else {
			l.Debug("participant found")
		}

		err = updateFn(ctx, prt)
		if err != nil {
			return err
		}

		return r.upsertParticipant(ctx, tx, prt)
	})
}

func (r *PostgresParticipantRepository) BotThreads(ctx context.Context, botId bots.BotId) ([]bots.BotThread, error) {
	var res []bots.BotThread
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		res, err = r.selectBotThreads(ctx, tx, botId)
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
	id bots.ParticipantId,
) (*bots.Participant, bool, error) {
	botId := string(id.BotId())
	userId := int64(id.UserId())

	row, err := r.getParticipantRow(ctx, qc, botId, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	threads, err := r.selectThreads(ctx, qc, id)
	if err != nil {
		return nil, false, err
	}

	prt, err := bots.UnmarshallParticipant(row.BotId, row.UserId, threads, row.CThread)
	if err != nil {
		return nil, false, err
	}

	return prt, true, nil
}

func (r *PostgresParticipantRepository) selectBotThreads(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botId bots.BotId,
) ([]bots.BotThread, error) {
	rows, err := r.selectBotThreadsRows(ctx, qc, string(botId))
	if err != nil {
		return nil, err
	}
	res := make([]bots.BotThread, len(rows))
	for i, row := range rows {
		answers, err := r.selectAnswers(ctx, qc, bots.ThreadId(row.Id))
		if err != nil {
			return nil, err
		}
		thread, err := bots.UnmarshallThread(row.Id, row.Key, row.State, answers, row.StartedAt)
		if err != nil {
			return nil, err
		}
		res[i] = bots.NewUserThread(thread, bots.UserId(row.UserId))
	}
	return res, nil
}

func (r *PostgresParticipantRepository) selectThreads(
	ctx context.Context,
	qc sqlx.QueryerContext,
	prtId bots.ParticipantId,
) ([]*bots.Thread, error) {
	botId := string(prtId.BotId())
	userId := int64(prtId.UserId())

	rows, err := r.selectThreadRows(ctx, qc, botId, userId)
	if err != nil {
		return nil, err
	}
	res := make([]*bots.Thread, len(rows))
	for i, row := range rows {
		answers, err := r.selectAnswers(ctx, qc, bots.ThreadId(row.Id))
		if err != nil {
			return nil, err
		}
		thread, err := bots.UnmarshallThread(row.Id, row.Key, row.State, answers, row.StartedAt)
		if err != nil {
			return nil, err
		}
		res[i] = thread
	}
	return res, nil
}

func (r *PostgresParticipantRepository) selectAnswers(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadId bots.ThreadId,
) (map[bots.State]bots.Message, error) {
	rows, err := r.selectAnswerRows(ctx, qc, string(threadId))
	if err != nil {
		return nil, err
	}
	res := make(map[bots.State]bots.Message)
	for _, row := range rows {
		msg, err := bots.NewMessage(row.Text)
		if err != nil {
			return nil, err
		}
		res[bots.State(row.State)] = msg
	}
	return res, nil
}

func (r *PostgresParticipantRepository) upsertParticipant(
	ctx context.Context,
	ec sqlx.ExtContext,
	prt *bots.Participant,
) error {
	botId := prt.Id().BotId()
	userId := prt.Id().UserId()
	prtRow := participantToRow(prt)
	threads := prt.Threads()
	threadRows := threadsToRows(botId, userId, threads)

	if err := r.upsertParticipantRow(ctx, ec, prtRow); err != nil {
		return err
	}

	if err := r.syncThreadRows(ctx, ec, botId, userId, threadRows); err != nil {
		return err
	}

	for _, thread := range threads {
		threadId := thread.Id()
		answerRows := answersToRows(threadId, thread.Answers())
		if err := r.syncAnswerRows(ctx, ec, threadId, answerRows); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresParticipantRepository) syncThreadRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botId bots.BotId,
	userId bots.UserId,
	rows []threadRow,
) error {
	const op = "PostgresParticipantRepository.syncThreadRows"
	l := r.log.With(
		slog.String("op", op),
		slog.String("bot_id", string(botId)),
		slog.Int64("user_id", int64(userId)),
	)
	l.Debug("syncing thread rows")

	dbRows, err := r.selectThreadRows(ctx, ec, string(botId), int64(userId))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, threadIdentity, diffcalc.Equal)
	l.Debug("calculated thread changes",
		slog.String("added", fmt.Sprintf("%v", changes.Added)),
		slog.String("updated", fmt.Sprintf("%v", changes.Updated)),
		slog.String("deleted", fmt.Sprintf("%v", changes.Deleted)),
	)

	if len(changes.Added) > 0 {
		err = r.insertThreadRows(ctx, ec, changes.Added)
		if err != nil {
			return err
		}
	}

	for _, row := range changes.Updated {
		err = r.updateThreadRow(ctx, ec, row)
		if err != nil {
			return err
		}
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteThreadRows(ctx, ec, changes.Deleted)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresParticipantRepository) syncAnswerRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	threadId bots.ThreadId,
	rows []answerRow,
) error {
	const op = "PostgresParticipantRepository.syncAnswerRows"
	l := r.log.With(
		slog.String("op", op),
		slog.String("thread_id", string(threadId)),
	)
	l.Debug("syncing answer rows")

	dbRows, err := r.selectAnswerRows(ctx, ec, string(threadId))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, answerIdentity, diffcalc.Equal)
	l.Debug("calculated answer changes",
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
	botId string,
	userId int64,
) (participantRow, error) {
	const op = "PostgresParticipantRepository.getParticipantRow"
	l := r.log.With(
		slog.String("op", op),
		slog.String("bot_id", botId),
		slog.Int64("user_id", userId),
	)

	l.Debug("querying participant row")
	var row participantRow
	err := pgutils.Get(ctx, qc, &row, `
		SELECT
			bot_id,
			user_id,
			cthread
		FROM participants
		WHERE
			bot_id = $1
			AND user_id = $2
		`,
		botId,
		userId,
	)
	if err != nil {
		l.Error("failed to get participant row", slog.String("error", err.Error()))
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
		slog.String("bot_id", row.BotId),
		slog.Int64("user_id", row.UserId),
	)

	l.Debug("upserting participant row")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		INSERT INTO
			participants (
				bot_id, 
				user_id,
				cthread
			)
		VALUES (
		    :bot_id,
			:user_id,
			:cthread
		)
		ON CONFLICT 
			(bot_id, user_id)
		DO UPDATE 
		SET
			cthread = :cthread
		`,
		row,
	))
	if err != nil {
		l.Error("failed to upsert participant row", slog.String("error", err.Error()))
		return fmt.Errorf("upserting participant row: %w", err)
	}
	return nil
}

func (r *PostgresParticipantRepository) selectThreadRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botId string,
	userId int64,
) ([]threadRow, error) {
	const op = "PostgresParticipantRepository.selectThreadRows"
	l := r.log.With(
		slog.String("op", op),
		slog.String("bot_id", botId),
		slog.Int64("user_id", userId),
	)

	l.Debug("querying thread rows")
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
			AND user_id = $2
		`,
		botId,
		userId,
	)
	if err != nil {
		l.Error("failed to query thread rows", slog.String("error", err.Error()))
		return nil, fmt.Errorf("selecting thread rows: %w", err)
	}
	return rows, nil
}

func (r *PostgresParticipantRepository) selectBotThreadsRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botId string,
) ([]threadRow, error) {
	const op = "PostgresParticipantRepository.selectBotThreadsRows"
	l := r.log.With(
		slog.String("op", op),
		slog.String("bot_id", botId),
	)

	l.Debug("querying bot thread rows")
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
		botId,
	)
	if err != nil {
		l.Error("failed to query bot thread rows", slog.String("error", err.Error()))
		return nil, fmt.Errorf("selecting bot thread rows: %w", err)
	}
	return rows, nil
}

func (r *PostgresParticipantRepository) insertThreadRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []threadRow,
) error {
	const op = "PostgresParticipantRepository.insertThreadRows"
	l := r.log.With(
		slog.String("op", op),
		slog.Int("rows", len(rows)),
	)

	l.Debug("inserting thread rows")
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
		`,
		rows,
	))
	if err != nil {
		l.Error("failed to insert thread rows", slog.String("error", err.Error()))
		return fmt.Errorf("inserting thread rows: %w", err)
	}
	return nil
}

func (r *PostgresParticipantRepository) updateThreadRow(
	ctx context.Context,
	ec sqlx.ExtContext,
	row threadRow,
) error {
	const op = "PostgresParticipantRepository.updateThreadRow"
	l := r.log.With(
		slog.String("op", op),
		slog.String("id", row.Id),
	)

	l.Debug("updating thread row")
	err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
		UPDATE threads
		SET
			bot_id     = :bot_id,
			user_id    = :user_id,
			key 	   = :key,
			state 	   = :state,
			started_at = :started_at
		WHERE
			id = :id
		`,
		row,
	))
	if err != nil {
		l.Error("failed to update thread row", slog.String("error", err.Error()))
		return fmt.Errorf("updating thread row: %w", err)
	}
	return nil
}

func (r *PostgresParticipantRepository) deleteThreadRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []threadRow,
) error {
	const op = "PostgresParticipantRepository.deleteThreadRows"
	l := r.log.With(
		slog.String("op", op),
		slog.Int("rows", len(rows)),
	)

	l.Debug("deleting answer rows")
	threadIds := threadRowsToIds(rows)
	err := pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		DELETE FROM threads
		WHERE
			id = ANY($1)`,
		pq.Array(threadIds),
	))
	if err != nil {
		l.Error("failed to delete thread rows", slog.String("error", err.Error()))
		return fmt.Errorf("deleting thread rows: %w", err)
	}
	return nil
}

func (r *PostgresParticipantRepository) selectAnswerRows(
	ctx context.Context,
	qc sqlx.QueryerContext,
	threadId string,
) ([]answerRow, error) {
	const op = "PostgresParticipantRepository.selectAnswerRows"
	l := r.log.With(
		slog.String("op", op),
		slog.String("thread_id", threadId),
	)

	l.Debug("querying answer rows")
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
		threadId,
	)
	if err != nil {
		l.Error("failed to query answer rows", slog.String("error", err.Error()))
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

	l.Debug("inserting answer rows")
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
		l.Error("failed to insert answer rows", slog.String("error", err.Error()))
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
		slog.String("thread_id", row.ThreadId),
		slog.Int("state", row.State),
	)

	l.Debug("updating answer row")
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
		l.Error("failed to update answer row", slog.String("error", err.Error()))
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

	l.Debug("deleting answer rows")
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
			l.Error("failed to delete answer rows", slog.String("error", err.Error()))
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
	// PK(BotId, UserId)
	BotId   string  `db:"bot_id"`
	UserId  int64   `db:"user_id"`
	CThread *string `db:"cthread"`
}

type threadRow struct {
	// PK(Id)
	Id        string    `db:"id"`
	BotId     string    `db:"bot_id"`
	UserId    int64     `db:"user_id"`
	Key       string    `db:"key"`
	State     int       `db:"state"`
	StartedAt time.Time `db:"started_at"`
}

func threadIdentity(lhs, rhs threadRow) bool {
	return lhs.Id == rhs.Id
}

type answerRow struct {
	// PK(ThreadId)
	ThreadId string `db:"thread_id"`
	State    int    `db:"state"`
	Text     string `db:"text"`
}

func answerIdentity(lhs, rhs answerRow) bool {
	return lhs.ThreadId == rhs.ThreadId && lhs.State == rhs.State
}

func participantToRow(prt *bots.Participant) participantRow {
	var cthread *string
	if ct, ok := prt.CurrentThread(); ok {
		s := string(ct.Id())
		cthread = &s
	}
	return participantRow{
		BotId:   string(prt.Id().BotId()),
		UserId:  int64(prt.Id().UserId()),
		CThread: cthread,
	}
}

func threadToRow(botId bots.BotId, userId bots.UserId, thread *bots.Thread) threadRow {
	return threadRow{
		Id:        string(thread.Id()),
		BotId:     string(botId),
		UserId:    int64(userId),
		Key:       string(thread.Key()),
		State:     int(thread.State()),
		StartedAt: thread.StartedAt(),
	}
}

func threadsToRows(botId bots.BotId, id bots.UserId, threads []*bots.Thread) []threadRow {
	res := make([]threadRow, len(threads))
	for i, thread := range threads {
		res[i] = threadToRow(botId, id, thread)
	}
	return res
}

func answerToRow(threadId bots.ThreadId, state bots.State, msg bots.Message) answerRow {
	return answerRow{
		ThreadId: string(threadId),
		State:    int(state),
		Text:     msg.Text(),
	}
}

func answersToRows(threadId bots.ThreadId, answers map[bots.State]bots.Message) []answerRow {
	res := make([]answerRow, 0, len(answers))
	for state, answer := range answers {
		res = append(res, answerToRow(threadId, state, answer))
	}
	return res
}

func threadRowsToIds(rows []threadRow) []string {
	res := make([]string, len(rows))
	for i, row := range rows {
		res[i] = row.Id
	}
	return res
}
