package postgres

import "github.com/bmstu-itstech/itsreg/internal/app/dto"

func botMetaFromRow(r botMetaRow) dto.BotMeta {
	return dto.BotMeta{
		ID:       r.ID,
		OwnerID:  r.OwnerID,
		ScriptID: r.ScriptID,
		Token:    r.Token,
		Deleted:  r.Deleted,
	}
}
