package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (r *Repository) OwnedMailing(ctx context.Context, mailingID bots.MailingID) (dto.OwnedMailing, error) {
	var m dto.OwnedMailing
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		row, err := r.getOwnedMailingRow(ctx, tx, mailingID.String())
		if errors.Is(err, sql.ErrNoRows) {
			return port.ErrMailingNotFound
		}
		if err != nil {
			return err
		}

		rRecs, err := r.selectMailingRecipientsByMailingIDRows(ctx, tx, mailingID.String())
		if err != nil {
			return err
		}

		rRes, err := r.selectMailingResultsByMailingIDRows(ctx, tx, mailingID.String())
		if err != nil {
			return err
		}

		m = dto.OwnedMailing{
			ID:           row.ID,
			OwnerID:      row.OwnerID,
			BotID:        row.BotID,
			Name:         row.Name,
			EntryKey:     row.EntryKey,
			Status:       row.Status,
			SuccessCount: row.SuccessCount,
			FailCount:    row.FailCount,
			PendingCount: row.PendingCount,
			TotalCount:   row.TotalCount,
			Recipients:   mailingRecipientsFromRowsToDTO(rRecs),
			Results:      mailingResultsFromRowsToDTO(rRes),
			CreatedAt:    row.CreatedAt,
			StartedAt:    row.StartedAt,
			CompletedAt:  row.CompletedAt,
		}
		return nil
	})
	return m, err
}
