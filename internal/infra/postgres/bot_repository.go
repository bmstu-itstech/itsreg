package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) Bot(ctx context.Context, id bots.BotID) (*bots.Bot, error) {
	rBot, err := r.getBotRow(ctx, r.db, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, port.ErrBotNotFound
	}
	if err != nil {
		return nil, err
	}
	return botFromRow(rBot)
}

func (r *Repository) BotsByOwnerID(ctx context.Context, ownerID bots.UserID) ([]*bots.Bot, error) {
	rBots, err := r.selectBotsByOwnerIDRows(ctx, r.db, ownerID.Int64())
	if err != nil {
		return nil, err
	}
	return botsFromRows(rBots)
}

func (r *Repository) SaveBot(ctx context.Context, bot *bots.Bot) error {
	rBot := botToRow(bot)
	err := r.insertBotRow(ctx, r.db, rBot)
	if pgutils.IsUniqueViolationError(err) {
		return port.ErrBotAlreadyExists
	}
	return err
}

func (r *Repository) UpdateBot(ctx context.Context, bot *bots.Bot) error {
	rBot := botToRow(bot)
	err := r.updateBotRow(ctx, r.db, rBot)
	if errors.Is(err, pgutils.ErrNoAffectedRows) {
		return port.ErrBotNotFound
	}
	return err
}
