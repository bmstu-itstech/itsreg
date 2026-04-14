package postgres

import (
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func answersFromRows(rs []answerRow) (map[bots.State]bots.Message, error) {
	res := make(map[bots.State]bots.Message, len(rs))
	for _, r := range rs {
		s, err := bots.NewState(r.State)
		if err != nil {
			return nil, err
		}
		m, err := bots.NewMessage(r.Text)
		if err != nil {
			return nil, err
		}
		res[s] = m
	}
	return res, nil
}

func threadToRow(t *bots.Thread) threadRow {
	return threadRow{
		ID:        t.ID().String(),
		BotID:     t.BotID().String(),
		UserID:    t.UserID().Int64(),
		Key:       t.Key().String(),
		State:     t.State().Int(),
		StartedAt: t.StartedAt(),
		UpdatedAt: t.StartedAt(),
	}
}

func answersToRows(as map[bots.State]bots.Message, threadID bots.ThreadID) []answerRow {
	res := make([]answerRow, 0, len(as))
	for s, a := range as {
		res = append(res, answerRow{
			ThreadID: threadID.String(),
			State:    s.Int(),
			Text:     a.String(),
		})
	}
	return res
}
