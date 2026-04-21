package postgres

import "github.com/bmstu-itstech/itsreg/internal/app/dto"

func scriptMetaFromRow(r scriptMetaRow) dto.ScriptMeta {
	return dto.ScriptMeta{
		ID:      r.ID,
		OwnerID: r.OwnerID,
		Desc:    r.Desc,
		Deleted: r.Deleted,
	}
}
