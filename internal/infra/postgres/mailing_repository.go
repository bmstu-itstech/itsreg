package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) Mailing(ctx context.Context, id bots.MailingID) (*bots.Mailing, error) {
	var mailing *bots.Mailing
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rMailing, err := r.getMailingRow(ctx, tx, id.String())
		if errors.Is(err, sql.ErrNoRows) {
			return port.ErrMailingNotFound
		}
		if err != nil {
			return fmt.Errorf("getMailingRow: %w", err)
		}

		rRecs, err := r.selectMailingRecipientsByMailingIDRows(ctx, tx, rMailing.ID)
		if err != nil {
			return fmt.Errorf("selectMailingResultsByMailingIDRows: %w", err)
		}

		rRes, err := r.selectMailingResultsByMailingIDRows(ctx, tx, rMailing.ID)
		if err != nil {
			return fmt.Errorf("selectMailingResultsByMailingIDRows: %w", err)
		}

		recipients := mailingRecipientsFromRows(rRecs)
		results, err := mailingResultsFromRows(rRes)
		if err != nil {
			return fmt.Errorf("mailingResultsFromRows: %w", err)
		}

		status, err := bots.MailingStatusFromString(rMailing.Status)
		if err != nil {
			return err
		}

		mailing, err = bots.RestoreMailing(
			bots.MailingID(rMailing.ID),
			bots.BotID(rMailing.BotID),
			rMailing.Name,
			bots.EntryKey(rMailing.EntryKey),
			status,
			recipients,
			results,
			rMailing.CreatedAt,
			rMailing.StartedAt,
			rMailing.CompletedAt,
		)
		return err
	})
	return mailing, err
}

func (r *Repository) MailingsByOwnerID(
	ctx context.Context, ownerID bots.UserID, filter port.MailingsFilter,
) ([]*bots.Mailing, error) {
	var mailings []*bots.Mailing
	err := pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rMailings, err := r.selectMailingsByOwnerIDRows(ctx, tx, ownerID.Int64(), selectMailingsFilterToDB(filter))
		if err != nil {
			return fmt.Errorf("selectMailingsByOwnerIDRows: %w", err)
		}

		ids := mailingRowsToIDs(rMailings)

		rRecs, err := r.selectMailingRecipientsByMailingIDsRows(ctx, tx, ids)
		if err != nil {
			return fmt.Errorf("selectMailingRecipientsByMailingIDsRows: %w", err)
		}

		rRes, err := r.selectMailingResultsByMailingIDsRows(ctx, tx, ids)
		if err != nil {
			return fmt.Errorf("selectMailingResultsByMailingIDsRows: %w", err)
		}

		recipientsByMailingID := hlpr.GroupBy(rRecs, func(rec mailingRecipientRow) string { return rec.MailingID })
		resultsByMailingID := hlpr.GroupBy(rRes, func(res mailingResultRow) string { return res.MailingID })

		for _, rMailing := range rMailings {
			recipients := mailingRecipientsFromRows(recipientsByMailingID[rMailing.ID])
			results, err2 := mailingResultsFromRows(resultsByMailingID[rMailing.ID])
			if err2 != nil {
				return fmt.Errorf("mailingResultsFromRows: %w", err2)
			}

			status, err2 := bots.MailingStatusFromString(rMailing.Status)
			if err2 != nil {
				return err2
			}

			mailing, err2 := bots.RestoreMailing(
				bots.MailingID(rMailing.ID),
				bots.BotID(rMailing.BotID),
				rMailing.Name,
				bots.EntryKey(rMailing.EntryKey),
				status,
				recipients,
				results,
				rMailing.CreatedAt,
				rMailing.StartedAt,
				rMailing.CompletedAt,
			)
			if err2 != nil {
				return err2
			}
			mailings = append(mailings, mailing)
		}

		return nil
	})
	return mailings, err
}

func (r *Repository) SaveMailing(ctx context.Context, m *bots.Mailing) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rMailing := mailingToRow(m)
		rRecs := mailingRecipientsToRows(m.Recipients(), m.ID())
		rRes := mailingResultsToRows(m.Results(), m.ID())

		err := r.insertMailingRow(ctx, tx, rMailing)
		if pgutils.IsUniqueViolationError(err) {
			return port.ErrMailingAlreadyExists
		}
		if err != nil {
			return fmt.Errorf("insertMailingRow: %w", err)
		}

		err = r.upsertMailingRecipientsRows(ctx, tx, rRecs)
		if err != nil {
			return fmt.Errorf("upsertMailingRecipientsRows: %w", err)
		}

		err = r.upsertMailingResultsRows(ctx, tx, rRes)
		if err != nil {
			return fmt.Errorf("upsertMailingResultsRows: %w", err)
		}

		return nil
	})
}

func (r *Repository) UpdateMailing(ctx context.Context, m *bots.Mailing) error {
	return pgutils.RunTx(ctx, r.db, func(tx *sqlx.Tx) error {
		rMailing := mailingToRow(m)
		rRecs := mailingRecipientsToRows(m.Recipients(), m.ID())
		rRes := mailingResultsToRows(m.Results(), m.ID())

		err := r.updateMailingRow(ctx, tx, rMailing)
		if errors.Is(err, pgutils.ErrNoAffectedRows) {
			return port.ErrMailingNotFound
		}
		if err != nil {
			return fmt.Errorf("updateMailingRow: %w", err)
		}

		err = r.upsertMailingRecipientsRows(ctx, tx, rRecs)
		if err != nil {
			return fmt.Errorf("upsertMailingRecipientsRows: %w", err)
		}

		err = r.upsertMailingResultsRows(ctx, tx, rRes)
		if err != nil {
			return fmt.Errorf("upsertMailingResultsRows: %w", err)
		}

		return nil
	})
}

func mailingRowsToIDs(rows []mailingRow) []string {
	res := make([]string, len(rows))
	for i, row := range rows {
		res[i] = row.ID
	}
	return res
}
