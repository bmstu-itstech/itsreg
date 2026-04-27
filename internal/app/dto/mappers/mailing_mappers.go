package mappers

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func MailingToDTO(m *bots.Mailing) dto.Mailing {
	return dto.Mailing{
		ID:           m.ID().String(),
		BotID:        m.BotID().String(),
		Name:         m.Name(),
		EntryKey:     m.EntryKey().String(),
		Status:       m.Status().String(),
		SuccessCount: m.SuccessCount(),
		FailCount:    m.FailCount(),
		PendingCount: m.PendingCount(),
		TotalCount:   m.RecipientsTotal(),
		CreatedAt:    m.CreatedAt(),
		StartedAt:    m.StartedAt(),
		CompletedAt:  m.CompletedAt(),
	}
}

func MailingsToDTO(ms []*bots.Mailing) []dto.Mailing {
	res := make([]dto.Mailing, len(ms))
	for i, m := range ms {
		res[i] = MailingToDTO(m)
	}
	return res
}

func UserIDsFromDTO(ds []int64) []bots.UserID {
	res := make([]bots.UserID, len(ds))
	for i, d := range ds {
		res[i] = bots.UserID(d)
	}
	return res
}
