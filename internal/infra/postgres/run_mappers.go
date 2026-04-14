package postgres

import "github.com/bmstu-itstech/itsreg/internal/domain/bots"

func runToRow(r *bots.Run) runRow {
	return runRow{
		ID:        r.ID().String(),
		BotID:     r.BotID().String(),
		Status:    r.Status().String(),
		ErrorMsg:  r.ErrorMsg(),
		StartedAt: r.StartedAt(),
		StoppedAt: r.StoppedAt(),
	}
}

func runFromRow(r runRow) (*bots.Run, error) {
	s, err := bots.StatusFromString(r.Status)
	if err != nil {
		return nil, err
	}
	return bots.RestoreRun(
		bots.RunID(r.ID),
		bots.BotID(r.BotID),
		s,
		r.ErrorMsg,
		r.StartedAt,
		r.StoppedAt,
	)
}

func runsFromRows(rs []runRow) ([]*bots.Run, error) {
	res := make([]*bots.Run, len(rs))
	for i, r := range rs {
		run, err := runFromRow(r)
		if err != nil {
			return nil, err
		}
		res[i] = run
	}
	return res, nil
}
