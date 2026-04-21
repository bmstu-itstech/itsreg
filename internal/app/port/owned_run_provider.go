package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type OwnedRunProvider interface {
	OwnedRun(ctx context.Context, id bots.RunID) (dto.OwnedRun, error)
}
