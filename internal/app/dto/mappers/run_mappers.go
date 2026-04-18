package mappers

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func RunToDTO(r *bots.Run) dto.Run {
	return dto.Run{
		ID:        r.ID().String(),
		BotID:     r.BotID().String(),
		Status:    r.Status().String(),
		ErrorMsg:  r.ErrorMsg(),
		StartedAt: r.StartedAt(),
		StoppedAt: r.StoppedAt(),
	}
}

func RunsToDTO(rs []*bots.Run) []dto.Run {
	res := make([]dto.Run, len(rs))
	for i, r := range rs {
		res[i] = RunToDTO(r)
	}
	return res
}
