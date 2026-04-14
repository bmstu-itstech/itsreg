package postgres

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) ThreadsTable(ctx context.Context, botID bots.BotID) (dto.ThreadsTable, error) {
	rows, err := r.selectThreadTableRows(ctx, r.db, botID.String())
	if err != nil {
		return dto.ThreadsTable{}, err
	}
	return groupThreadsAnswers(rows), nil
}
