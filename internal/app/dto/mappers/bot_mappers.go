package mappers

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func BotToDTO(b *bots.Bot) dto.Bot {
	return dto.Bot{
		ID:        b.ID().String(),
		OwnerID:   b.OwnerID().Int64(),
		ScriptID:  b.ScriptID().String(),
		Desc:      b.Desc(),
		CreatedAt: b.CreatedAt(),
		UpdatedAt: b.UpdatedAt(),
	}
}

func BotsToDTO(bs []*bots.Bot) []dto.Bot {
	res := make([]dto.Bot, len(bs))
	for i, b := range bs {
		res[i] = BotToDTO(b)
	}
	return res
}
