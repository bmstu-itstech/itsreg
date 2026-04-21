package postgres

import "github.com/bmstu-itstech/itsreg/internal/app/dto"

func ownedRunFromRow(r ownedRunRow) dto.OwnedRun {
	return dto.OwnedRun{
		ID:        r.ID,
		OwnerID:   r.OwnerID,
		BotID:     r.BotID,
		Status:    r.Status,
		ErrorMsg:  r.ErrorMsg,
		StartedAt: r.StartedAt,
		StoppedAt: r.StoppedAt,
	}
}
