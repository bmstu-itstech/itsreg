package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) BotMeta(ctx context.Context, id bots.BotID) (dto.BotMeta, error) {
	row, err := r.getBotMetaRow(ctx, r.db, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return dto.BotMeta{}, port.ErrBotNotFound
	}
	if err != nil {
		return dto.BotMeta{}, err
	}
	return botMetaFromRow(row), nil
}
