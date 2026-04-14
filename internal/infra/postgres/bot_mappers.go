package postgres

import "github.com/bmstu-itstech/itsreg/internal/domain/bots"

func botToRow(b *bots.Bot) botRow {
	return botRow{
		ID:        b.ID().String(),
		OwnerID:   b.OwnerID().Int64(),
		ScriptID:  b.ScriptID().String(),
		Token:     b.Token().String(),
		Desc:      b.Desc(),
		CreatedAt: b.CreatedAt(),
		UpdatedAt: b.UpdatedAt(),
		DeletedAt: b.DeletedAt(),
	}
}

func botFromRow(r botRow) (*bots.Bot, error) {
	return bots.RestoreBot(
		bots.BotID(r.ID),
		bots.UserID(r.OwnerID),
		bots.ScriptID(r.ScriptID),
		bots.Token(r.Token),
		r.Desc,
		r.CreatedAt,
		r.UpdatedAt,
		r.DeletedAt,
	)
}

func botsFromRows(rs []botRow) ([]*bots.Bot, error) {
	bs := make([]*bots.Bot, len(rs))
	for i, r := range rs {
		b, err := botFromRow(r)
		if err != nil {
			return nil, err
		}
		bs[i] = b
	}
	return bs, nil
}
