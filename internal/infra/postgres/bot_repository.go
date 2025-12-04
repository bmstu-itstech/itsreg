package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
)

func (r *Repository) Bot(ctx context.Context, id bots.BotID) (*bots.Bot, error) {
	var bot *bots.Bot
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		bot, err = r.getBot(ctx, tx, id)
		return err
	})
	return bot, err
}

func (r *Repository) UserBots(ctx context.Context, author bots.UserID) ([]*bots.Bot, error) {
	var _bots []*bots.Bot
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		_bots, err = r.selectBotsByAuthor(ctx, tx, author)
		return err
	})
	return _bots, err
}

func (r *Repository) EnabledBots(ctx context.Context) ([]*bots.Bot, error) {
	var _bots []*bots.Bot
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		var err error
		_bots, err = r.selectEnabledBots(ctx, tx)
		return err
	})
	return _bots, err
}

func (r *Repository) UpsertBot(ctx context.Context, bot *bots.Bot) error {
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

func (r *Repository) DeleteBot(ctx context.Context, id bots.BotID) error {
	err := r.softDeleteBotRow(ctx, r.db, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s", port.ErrBotNotFound, string(id))
	}
	return err
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

func (r *Repository) selectBotsByAuthor(
	ctx context.Context,
	qc sqlx.QueryerContext,
	author bots.UserID,
) ([]*bots.Bot, error) {
	rows, err := r.selectBotRowsByAuthor(ctx, qc, int64(author))
	if err != nil {
		return nil, err
	}
	res := make([]*bots.Bot, len(rows))
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
		bot, err2 := bots.UnmarshallBot(
			row.ID, row.Token, row.Author, row.Enabled, script, row.CreatedAt.In(time.Local),
		)
		if err2 != nil {
			return nil, err2
		}
		res[i] = bot
	}
	return res, nil
}

func (r *Repository) selectEnabledBots(
	ctx context.Context,
	qc sqlx.QueryerContext,
) ([]*bots.Bot, error) {
	rows, err := r.selectEnabledBotRows(ctx, qc)
	if err != nil {
		return nil, err
	}
	res := make([]*bots.Bot, len(rows))
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
		bot, err2 := bots.UnmarshallBot(
			row.ID, row.Token, row.Author, row.Enabled, script, row.CreatedAt.In(time.Local),
		)
		if err2 != nil {
			return nil, err2
		}
		res[i] = bot
	}
	return res, nil
}

func (r *Repository) getBot(
	ctx context.Context,
	qc sqlx.QueryerContext,
	id bots.BotID,
) (*bots.Bot, error) {
	row, err := r.getBotRow(ctx, qc, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s", port.ErrBotNotFound, id)
	} else if err != nil {
		return nil, err
	}
	entries, err := r.selectEntries(ctx, qc, id)
	if err != nil {
		return nil, err
	}
	nodes, err := r.selectNodes(ctx, qc, id)
	if err != nil {
		return nil, err
	}
	script, err := bots.NewScript(nodes, entries)
	if err != nil {
		return nil, err
	}
	return bots.UnmarshallBot(row.ID, row.Token, row.Author, row.Enabled, script, row.CreatedAt.In(time.Local))
}

func (r *Repository) selectEntries(
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
		state, err2 := bots.NewState(row.Start)
		if err2 != nil {
			return nil, err2
		}
		entry, err2 := bots.NewEntry(bots.EntryKey(row.Key), state)
		if err2 != nil {
			return nil, err2
		}
		res[i] = entry
	}
	return res, nil
}

func (r *Repository) selectNodes(
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
		state, err2 := bots.NewState(row.State)
		if err2 != nil {
			return nil, err2
		}
		edges, err2 := r.selectEdges(ctx, qc, botID, state)
		if err2 != nil {
			return nil, err2
		}
		msgs, err2 := r.selectMessages(ctx, qc, botID, state)
		if err2 != nil {
			return nil, err2
		}
		opts, err2 := r.selectOptions(ctx, qc, botID, state)
		if err2 != nil {
			return nil, err2
		}
		node, err2 := bots.NewNode(state, row.Title, edges, msgs, opts)
		if err2 != nil {
			return nil, err2
		}
		res[i] = node
	}
	return res, nil
}

func (r *Repository) selectEdges(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Edge, error) {
	rows, err := r.selectEdgeRows(ctx, qc, string(botID), state.Int())
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
		to, err2 := bots.NewState(row.ToState)
		if err2 != nil {
			return nil, err2
		}
		edge := bots.NewEdge(pred, to, oper)
		res[i] = edge
	}
	return res, nil
}

func (r *Repository) selectMessages(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Message, error) {
	rows, err := r.selectMessageRows(ctx, qc, string(botID), state.Int())
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

func (r *Repository) selectOptions(
	ctx context.Context,
	qc sqlx.QueryerContext,
	botID bots.BotID,
	state bots.State,
) ([]bots.Option, error) {
	rows, err := r.selectOptionRows(ctx, qc, string(botID), state.Int())
	if err != nil {
		return nil, err
	}
	res := make([]bots.Option, len(rows))
	for i, row := range rows {
		o, err2 := bots.NewOption(row.Text)
		if err2 != nil {
			return nil, err2
		}
		res[i] = o
	}
	return res, nil
}

func (r *Repository) syncEntryRows(
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

func (r *Repository) syncNodeRows(
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

func (r *Repository) syncEdgeRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []edgeRow,
) error {
	dbRows, err := r.selectEdgeRows(ctx, ec, string(botID), state.Int())
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[edgeRow], diffcalc.Equal[edgeRow])

	if changes.IsZero() {
		return nil
	}

	if len(dbRows) > 0 {
		err = r.deleteEdgeRows(ctx, ec, string(botID), state.Int())
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

func (r *Repository) syncMessageRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []messageRow,
) error {
	dbRows, err := r.selectMessageRows(ctx, ec, string(botID), state.Int())
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[messageRow], diffcalc.Equal[messageRow])

	if changes.IsZero() {
		return nil
	}

	if len(dbRows) > 0 {
		err = r.deleteMessageRows(ctx, ec, string(botID), state.Int())
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

func (r *Repository) syncOptionRows(
	ctx context.Context,
	ec sqlx.ExtContext,
	botID bots.BotID,
	state bots.State,
	rows []optionRow,
) error {
	dbRows, err := r.selectOptionRows(ctx, ec, string(botID), state.Int())
	if err != nil {
		return err
	}

	changes := diffcalc.Changes(dbRows, rows, diffcalc.Equal[optionRow], diffcalc.Equal[optionRow])

	if changes.IsZero() {
		return nil
	}

	if len(changes.Deleted) > 0 {
		err = r.deleteOptionRows(ctx, ec, string(botID), state.Int())
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
