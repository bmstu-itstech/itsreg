package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

type PostgresBotRepository struct {
	db *sqlx.DB
}

func NewPostgresBotRepository(db *sqlx.DB) *PostgresBotRepository {
	return &PostgresBotRepository{
		db: db,
	}
}

func (r *PostgresBotRepository) Bot(ctx context.Context, id bots.BotID) (bots.Bot, error) {
	var bot bots.Bot
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		bot, err = r.getBot(ctx, tx, id)
		return err
	})
	return bot, err
}

func (r *PostgresBotRepository) UserBots(ctx context.Context, author bots.UserID) ([]bots.Bot, error) {
	var _bots []bots.Bot
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		_bots, err = r.selectBotsByAuthor(ctx, tx, author)
		return err
	})
	return _bots, err
}

func (r *PostgresBotRepository) Upsert(ctx context.Context, bot bots.Bot) error {
	_botRow := botToRow(bot)
	entryRows := entriesToRows(bot.ID(), bot.Script().Entries())
	nodes := bot.Script().Nodes()
	nodeRows := nodesToRows(bot.ID(), nodes)

	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		if err := r.upsertBotRow(ctx, tx, _botRow); err != nil {
			return err
		}
		if err := r.syncNodeRows(ctx, tx, bot.ID(), nodeRows); err != nil {
			return err
		}
		if err := r.syncEntryRows(ctx, tx, bot.ID(), entryRows); err != nil {
			return err
		}
		for _, node := range nodes {
			edgeRows := edgesToRows(bot.ID(), node.State(), node.Edges())
			if err := r.syncEdgeRows(ctx, tx, bot.ID(), node.State(), edgeRows); err != nil {
				return err
			}
			messageRows := messagesToRows(bot.ID(), node.State(), node.Messages())
			if err := r.syncMessageRows(ctx, tx, bot.ID(), node.State(), messageRows); err != nil {
				return err
			}
			optionRows := optionsToRows(bot.ID(), node.State(), node.Options())
			if err := r.syncOptionRows(ctx, tx, bot.ID(), node.State(), optionRows); err != nil {
				return err
			}
		}
		return nil
	})
}

//
//
// ОПЕРАЦИИ НАД СУЩНОСТЯМИ ВНУТРИ АГГРЕГАТА
//
//

// При работе с ботом в БД возникает ряд сложностей.
// 1. Бот является агрегатом, то есть содержит в себе множество разнотипных сущностей со своими идентификаторами;
// 2. При обновлении бота нельзя стереть его полностью и записать заново (как было в v1), так как в таком случае
//    даже простое изменение текста в одном из узлов сопровождается удалением всех ответов из-за каскадного удаления.
// Следовательно, нужно точечно обновлять бота. Для этого:
// 1. Бот переводится в промежуточное состояние между БД и доменной моделью. Это набор всех строк всех таблиц,
//    связанных с ботом. Так, это строка таблицы bots, строки nodes, entries, edges и тому подобное.
// 2. Аналогичное промежуточное состояние воссоздаётся из БД путём SELECT запросов.
// 3. Так как промежуточное состояние состоит из набора простых структур, то можно проводить их сравнение по значению.
// 4. Некоторые объекты внутри бота имеют идентичность (являются сущностями), поэтому можно проследить их изменение.
// 5. Далее для сущностей и объектов-значений (далее - ОЗ) составляются таблицы изменений.
//    5.1. Для сущностей таблица содержит добавленные строки (такого ID нет в текущем промежуточном состоянии бота),
//    	   удалённые строки (такого ID нет в ожидаемом промежуточном состоянии бота), обновлённые строки
//         (ID есть и там, и там, но строки не равны). Остальные строки игнорируются.
//    5.2. Так как ОЗ не имеют своей идентичности, то понятия "обновления" для них не существует (ОЗ неизменяемы).
//		   Из этого так же следует, что, как правило, к ОЗ нельзя обращаться точечно, поэтому при анализе изменений
//		   строк важен только факт их изменения.
// 6. В соответствии с таблицей выполняются изменения в БД.
//	  6.1. Для сущностей происходят точечные добавления, изменения и удаления строк из БД.
//	  6.2. Для объектов-значений происходит полное удаление всех ОЗ, связанных с ближайшим родителем-сущностью (чтобы
//	       охватить минимальное количество строк), а потом добавление их заново. Вынужденная мера, так как ОЗ не имеют
//	       своих ID.
//
// Профит! Как можно улучшить?
// 1. Попробовать уменьшить количество ОЗ (а оно надо?).
// 2. Сделать PK для объектов-значений совокупностью всех столбцов, чтобы работать с ними так же, как с сущностями
//    (а оно надо?).
//
// Почему оно не надо? Потому что главное, что теперь не происходит никаких лишних операций с теми строками, на
// которые потенциально могут ссылаться другие (например, answers на nodes). А так как ОЗ не имеют своих ID,
// то и ссылаться на них некому.

// Функции типа sync выполняют необходимую синхронизацию строк в БД и желаемого состояния, передаваемого в аргументах.

func (r *PostgresBotRepository) selectBotsByAuthor(
	ctx context.Context,
	qc sqlx.QueryerContext,
	author bots.UserID,
) ([]bots.Bot, error) {
	rows, err := r.selectBotRowsByAuthor(ctx, qc, int64(author))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Bot, len(rows))
	for i, row := range rows {
		entries, err2 := r.selectEntries(ctx, qc, bots.BotID(row.ID))
		if err2 != nil {
			return nil, err2
		}
		nodes, err2 := r.selectNodes(ctx, qc, bots.BotID(row.ID))
		if err2 != nil {
			return nil, err2
		}
		script, err2 := bots.NewScript(nodes, entries)
		if err2 != nil {
			return nil, err2
		}
		bot, err2 := bots.UnmarshallBot(row.ID, row.Token, row.Author, script, row.CreatedAt.In(time.Local))
		if err2 != nil {
			return nil, err2
		}
		res[i] = bot
	}
	return res, nil
}

func (r *PostgresBotRepository) getBot(
	ctx context.Context,
	qc sqlx.QueryerContext,
	id bots.BotID,
) (bots.Bot, error) {
	row, err := r.getBotRow(ctx, qc, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return bots.Bot{}, fmt.Errorf("%w: %s", bots.ErrBotNotFound, id)
	} else if err != nil {
		return bots.Bot{}, err
	}
	entries, err := r.selectEntries(ctx, qc, id)
	if err != nil {
		return bots.Bot{}, err
	}
	nodes, err := r.selectNodes(ctx, qc, id)
	if err != nil {
		return bots.Bot{}, err
	}
	script, err := bots.NewScript(nodes, entries)
	if err != nil {
		return bots.Bot{}, err
	}
	return bots.UnmarshallBot(row.ID, row.Token, row.Author, script, row.CreatedAt.In(time.Local))
}

func (r *PostgresBotRepository) selectEntries(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
) ([]bots.Entry, error) {
	rows, err := r.selectEntryRows(ctx, qc, string(botID))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Entry, len(rows))
	for i, row := range rows {
		entry, err2 := bots.NewEntry(bots.EntryKey(row.Key), bots.State(row.Start))
		if err2 != nil {
			return nil, err2
		}
		res[i] = entry
	}
	return res, nil
}

func (r *PostgresBotRepository) selectNodes(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
) ([]bots.Node, error) {
	rows, err := r.selectNodeRows(ctx, qc, string(botID))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Node, len(rows))
	for i, row := range rows {
		edges, err2 := r.selectEdges(ctx, qc, botID, bots.State(row.State))
		if err2 != nil {
			return nil, err2
		}
		msgs, err2 := r.selectMessages(ctx, qc, botID, bots.State(row.State))
		if err2 != nil {
			return nil, err2
		}
		opts, err2 := r.selectOptions(ctx, qc, botID, bots.State(row.State))
		if err2 != nil {
			return nil, err2
		}
		node, err2 := bots.NewNode(bots.State(row.State), row.Title, edges, msgs, opts)
		if err2 != nil {
			return nil, err2
		}
		res[i] = node
	}
	return res, nil
}

func (r *PostgresBotRepository) selectEdges(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Edge, error) {
	rows, err := r.selectEdgeRows(ctx, qc, string(botID), int(state))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Edge, len(rows))
	for i, row := range rows {
		pred, err2 := predicateFromStrings(row.PredType, row.PredData)
		if err2 != nil {
			return nil, err2
		}
		oper, err2 := operationFromString(row.Operation)
		if err2 != nil {
			return nil, err2
		}
		edge := bots.NewEdge(pred, bots.State(row.ToState), oper)
		res[i] = edge
	}
	return res, nil
}

func (r *PostgresBotRepository) selectMessages(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Message, error) {
	rows, err := r.selectMessageRows(ctx, qc, string(botID), int(state))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Message, len(rows))
	for i, row := range rows {
		res[i], err = bots.NewMessage(row.Text)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (r *PostgresBotRepository) selectOptions(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Option, error) {
	rows, err := r.selectOptionRows(ctx, qc, string(botID), int(state))
	if err != nil {
		return nil, err
	}
	res := make([]bots.Option, len(rows))
	for i, row := range rows {
		res[i] = bots.Option(row.Text)
	}
	return res, nil
}

func (r *PostgresBotRepository) syncEntryRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	rows []entryRow,
) error {
	dbRows, err := r.selectEntryRows(ctx, ec, string(botID))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, entryIdentity, diffcalc.Equal)

	if len(changes.Added) > 0 {
		err = r.insertEntryRows(ctx, ec, changes.Added)
		if err != nil {
			return err
		}
	}

	for _, row := range changes.Updated {
		err = r.updateEntryRow(ctx, ec, row)
		if err != nil {
			return err
		}
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteEntryRows(ctx, ec, changes.Deleted)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresBotRepository) syncNodeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	rows []nodeRow,
) error {
	dbRows, err := r.selectNodeRows(ctx, ec, string(botID))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, nodeIdentity, diffcalc.Equal)

	if len(changes.Added) > 0 {
		err = r.insertNodeRows(ctx, ec, changes.Added)
		if err != nil {
			return err
		}
	}

	for _, row := range changes.Updated {
		err = r.updateNodeRow(ctx, ec, row)
		if err != nil {
			return err
		}
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteNodeRows(ctx, ec, changes.Deleted)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresBotRepository) syncEdgeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []edgeRow,
) error {
	dbRows, err := r.selectEdgeRows(ctx, ec, string(botID), int(state))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[edgeRow], diffcalc.Equal[edgeRow])

	if changes.IsZero() {
		return nil
	}

	if len(dbRows) > 0 {
		err = r.deleteEdgeRows(ctx, ec, string(botID), int(state))
		if err != nil {
			return err
		}
	}

	if len(rows) > 0 {
		err = r.insertEdgeRows(ctx, ec, rows)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresBotRepository) syncMessageRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []messageRow,
) error {
	dbRows, err := r.selectMessageRows(ctx, ec, string(botID), int(state))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[messageRow], diffcalc.Equal[messageRow])

	if changes.IsZero() {
		return nil
	}

	if len(dbRows) > 0 {
		err = r.deleteMessageRows(ctx, ec, string(botID), int(state))
		if err != nil {
			return err
		}
	}

	if len(rows) > 0 {
		err = r.insertMessageRows(ctx, ec, rows)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresBotRepository) syncOptionRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []optionRow,
) error {
	dbRows, err := r.selectOptionRows(ctx, ec, string(botID), int(state))
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[optionRow], diffcalc.Equal[optionRow])

	if changes.IsZero() {
		return nil
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteOptionRows(ctx, ec, string(botID), int(state))
		if err != nil {
			return err
		}
	}

	if len(rows) > 0 {
		err = r.insertOptionRows(ctx, ec, rows)
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

// CRUD интерфейс для всех строк всех таблиц, связанных с ботом.

func (r *PostgresBotRepository) getBotRow(
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
			created_at
		FROM bots
		WHERE
			id = $1
		`,
		botID,
	)
	if err != nil {
		return row, fmt.Errorf("selecting bot row: %w", err)
	}
	return row, nil
}

func (r *PostgresBotRepository) selectBotRowsByAuthor(
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
			created_at
		FROM bots
		WHERE
			author = $1
		`,
		author,
	)
	if err != nil {
		return nil, fmt.Errorf("selecting bot rows by author: %w", err)
	}
	return rows, nil
}

func (r *PostgresBotRepository) upsertBotRow(
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
				created_at
			)
		VALUES (
		    :id,
			:token,
			:author,
			:created_at
		)
		ON CONFLICT 
			(id)
		DO UPDATE 
		SET
			token      = :token,
			author     = :author,
			created_at = :created_at
		`,
		row,
	))
	if err != nil {
		return fmt.Errorf("upserting bot row: %w", err)
	}
	return nil
}

func (r *PostgresBotRepository) selectEntryRows(
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

func (r *PostgresBotRepository) insertEntryRows(
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
			:key,
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

func (r *PostgresBotRepository) updateEntryRow(
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
			AND key = :key
		`,
		row,
	))
	if err != nil {
		return fmt.Errorf("update entry rows: %w", err)
	}
	return nil
}

func (r *PostgresBotRepository) deleteEntryRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	rows []entryRow,
) error {
	for _, row := range rows {
		err := pgutils.RequireAffected(pgutils.NamedExec(ctx, ec, `
			DELETE FROM entries
			WHERE
				bot_id = :bot_id
				AND key = :key
			`,
			row,
		))
		if err != nil {
			return fmt.Errorf("deleting entry rows: %w", err)
		}
	}
	return nil
}

func (r *PostgresBotRepository) selectNodeRows(
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

func (r *PostgresBotRepository) insertNodeRows(
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

func (r *PostgresBotRepository) updateNodeRow(
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

func (r *PostgresBotRepository) deleteNodeRows(
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

func (r *PostgresBotRepository) selectEdgeRows(
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

func (r *PostgresBotRepository) insertEdgeRows(
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

func (r *PostgresBotRepository) deleteEdgeRows(
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

func (r *PostgresBotRepository) selectMessageRows(
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

func (r *PostgresBotRepository) insertMessageRows(
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

func (r *PostgresBotRepository) deleteMessageRows(
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

func (r *PostgresBotRepository) selectOptionRows(
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

func (r *PostgresBotRepository) insertOptionRows(
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

func (r *PostgresBotRepository) deleteOptionRows(
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

//
//
// МОДЕЛИ БАЗЫ ДАННЫХ
//
//

type botRow struct {
	// PK (ID)
	ID        string    `db:"id"`
	Token     string    `db:"token"`
	Author    int64     `db:"author"`
	CreatedAt time.Time `db:"created_at"`
}

type entryRow struct {
	// PK (BotID, Key)
	BotID string `db:"bot_id"`
	Key   string `db:"key"`
	Start int    `db:"start"`
}

func entryIdentity(lhs, rhs entryRow) bool {
	return lhs.BotID == rhs.BotID && lhs.Key == rhs.Key
}

type nodeRow struct {
	// PK (BotID, State)
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Title string `db:"title"`
}

func nodeIdentity(lhs, rhs nodeRow) bool {
	return lhs.BotID == rhs.BotID && lhs.State == rhs.State
}

type edgeRow struct {
	BotID     string `db:"bot_id"`
	State     int    `db:"state"`
	ToState   int    `db:"to_state"`
	Operation string `db:"operation"`
	PredType  string `db:"pred_type"`
	PredData  string `db:"pred_data"`
}

type messageRow struct {
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Text  string `db:"text"`
}

type optionRow struct {
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Text  string `db:"text"`
}

func operationToString(op bots.Operation) string {
	switch op.(type) {
	case bots.NoOp:
		return "noop"
	case bots.SaveOp:
		return "save"
	case bots.AppendOp:
		return "append"
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func operationFromString(s string) (bots.Operation, error) {
	switch s {
	case "noop":
		return bots.NoOp{}, nil
	case "save":
		return bots.SaveOp{}, nil
	case "append":
		return bots.AppendOp{}, nil
	default:
		return nil, fmt.Errorf("invalid operation %s, expected one of ['noop', 'save', 'append']", s)
	}
}

func predicateToStrings(p bots.Predicate) (string, string) {
	switch p := p.(type) {
	case bots.AlwaysTruePredicate:
		return "always", ""
	case bots.ExactMatchPredicate:
		return "exact", p.Text()
	case bots.RegexMatchPredicate:
		return "regexp", p.Pattern()
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func predicateFromStrings(ptype string, pdata string) (bots.Predicate, error) {
	switch ptype {
	case "always":
		return bots.AlwaysTruePredicate{}, nil
	case "exact":
		return bots.NewExactMatchPredicate(pdata)
	case "regexp":
		return bots.NewRegexMatchPredicate(pdata)
	default:
		return nil, fmt.Errorf("invalid predicate type %s, expected one of ['always', 'exact', 'regexp']", ptype)
	}
}

func botToRow(bot bots.Bot) botRow {
	return botRow{
		ID:        string(bot.ID()),
		Token:     string(bot.Token()),
		Author:    int64(bot.Author()),
		CreatedAt: bot.CreatedAt().In(time.UTC),
	}
}

func entryToRow(botID bots.BotID, entry bots.Entry) entryRow {
	return entryRow{
		BotID: string(botID),
		Key:   string(entry.Key()),
		Start: int(entry.Start()),
	}
}

func entriesToRows(botID bots.BotID, entries []bots.Entry) []entryRow {
	res := make([]entryRow, len(entries))
	for i, entry := range entries {
		res[i] = entryToRow(botID, entry)
	}
	return res
}

func nodeToRow(botID bots.BotID, node bots.Node) nodeRow {
	return nodeRow{
		BotID: string(botID),
		State: int(node.State()),
		Title: node.Title(),
	}
}

func nodesToRows(botID bots.BotID, nodes []bots.Node) []nodeRow {
	res := make([]nodeRow, len(nodes))
	for i, node := range nodes {
		res[i] = nodeToRow(botID, node)
	}
	return res
}

func edgeToRow(botID bots.BotID, state bots.State, edge bots.Edge) edgeRow {
	ptype, pdata := predicateToStrings(edge.Predicate)
	return edgeRow{
		BotID:     string(botID),
		State:     int(state),
		ToState:   int(edge.To()),
		Operation: operationToString(edge.Operation()),
		PredType:  ptype,
		PredData:  pdata,
	}
}

func edgesToRows(botID bots.BotID, state bots.State, edges []bots.Edge) []edgeRow {
	res := make([]edgeRow, len(edges))
	for i, edge := range edges {
		res[i] = edgeToRow(botID, state, edge)
	}
	return res
}

func messageToRow(botID bots.BotID, state bots.State, msg bots.Message) messageRow {
	return messageRow{
		BotID: string(botID),
		State: int(state),
		Text:  msg.Text(),
	}
}

func messagesToRows(botID bots.BotID, state bots.State, msgs []bots.Message) []messageRow {
	res := make([]messageRow, len(msgs))
	for i, msg := range msgs {
		res[i] = messageToRow(botID, state, msg)
	}
	return res
}

func optionToRow(botID bots.BotID, state bots.State, opt bots.Option) optionRow {
	return optionRow{
		BotID: string(botID),
		State: int(state),
		Text:  string(opt),
	}
}

func optionsToRows(botID bots.BotID, state bots.State, opts []bots.Option) []optionRow {
	res := make([]optionRow, len(opts))
	for i, opt := range opts {
		res[i] = optionToRow(botID, state, opt)
	}
	return res
}
