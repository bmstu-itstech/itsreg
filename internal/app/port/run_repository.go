package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var (
	ErrRunNotFound            = errors.New("run not found")
	ErrActiveRunAlreadyExists = errors.New("active run already exists")
)

type RunsFilter struct {
	BotID  *bots.BotID
	Status *bots.Status
}

type RunRepository interface {
	Run(ctx context.Context, id bots.RunID) (*bots.Run, error)
	RunsByOwnerID(ctx context.Context, ownerID bots.UserID, filter RunsFilter) ([]*bots.Run, error)
	ActiveRuns(ctx context.Context) ([]*bots.Run, error)

	SaveRun(ctx context.Context, run *bots.Run) error
	UpdateRun(ctx context.Context, run *bots.Run) error
}
