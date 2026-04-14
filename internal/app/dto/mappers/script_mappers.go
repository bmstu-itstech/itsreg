package mappers

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func ScriptToDTO(s *bots.Script) dto.Script {
	return dto.Script{
		ID:        s.ID().String(),
		OwnerID:   s.OwnerID().Int64(),
		Desc:      s.Desc(),
		Nodes:     nodesToDTO(s.Nodes()),
		Entries:   entriesToDTO(s.Entries()),
		CreatedAt: s.CreatedAt(),
		UpdatedAt: s.UpdatedAt(),
	}
}

func ScriptsToDTO(ss []*bots.Script) []dto.Script {
	res := make([]dto.Script, len(ss))
	for i, s := range ss {
		res[i] = ScriptToDTO(s)
	}
	return res
}
