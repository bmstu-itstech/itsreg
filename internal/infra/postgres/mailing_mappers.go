package postgres

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
)

func selectMailingsFilterToDB(f port.MailingsFilter) selectMailingsFilter {
	var filter selectMailingsFilter
	if f.BotID != nil {
		filter.BotID = hlpr.Ptr(f.BotID.String())
	}
	if f.Status != nil {
		filter.Status = hlpr.Ptr(f.Status.String())
	}
	return filter
}

func mailingRecipientsFromRows(rows []mailingRecipientRow) []bots.UserID {
	res := make([]bots.UserID, len(rows))
	for i, row := range rows {
		res[i] = bots.UserID(row.UserID)
	}
	return res
}

func mailingRecipientsFromRowsToDTO(rows []mailingRecipientRow) []int64 {
	res := make([]int64, len(rows))
	for i, row := range rows {
		res[i] = row.UserID
	}
	return res
}

func mailingResultsFromRows(rows []mailingResultRow) ([]bots.UserMailingResult, error) {
	res := make([]bots.UserMailingResult, len(rows))
	for i, row := range rows {
		r, err := bots.RestoreUserMailingResult(
			bots.UserID(row.UserID),
			row.Success,
			row.ErrorMsg,
		)
		if err != nil {
			return nil, err
		}
		res[i] = r
	}
	return res, nil
}

func mailingResultsFromRowsToDTO(rows []mailingResultRow) []dto.UserMailingResult {
	res := make([]dto.UserMailingResult, len(rows))
	for i, row := range rows {
		res[i] = dto.UserMailingResult{
			UserID:   row.UserID,
			Success:  row.Success,
			ErrorMsg: row.ErrorMsg,
		}
	}
	return res
}

func mailingToRow(m *bots.Mailing) mailingRow {
	return mailingRow{
		ID:          m.ID().String(),
		BotID:       m.BotID().String(),
		Name:        m.Name(),
		EntryKey:    m.EntryKey().String(),
		Status:      m.Status().String(),
		CreatedAt:   m.CreatedAt(),
		StartedAt:   m.StartedAt(),
		CompletedAt: m.CompletedAt(),
	}
}

func mailingRecipientsToRows(rs []bots.UserID, mailingID bots.MailingID) []mailingRecipientRow {
	res := make([]mailingRecipientRow, len(rs))
	for i, r := range rs {
		res[i] = mailingRecipientRow{
			MailingID: mailingID.String(),
			UserID:    r.Int64(),
		}
	}
	return res
}

func mailingResultsToRows(results []bots.UserMailingResult, mailingID bots.MailingID) []mailingResultRow {
	res := make([]mailingResultRow, len(results))
	for i, r := range results {
		res[i] = mailingResultRow{
			MailingID: mailingID.String(),
			UserID:    r.UserID().Int64(),
			Success:   r.Success(),
			ErrorMsg:  r.ErrorMessage(),
		}
	}
	return res
}
