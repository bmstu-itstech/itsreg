package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type ThreadsTableProvider interface {
	ThreadsTable(ctx context.Context, botID bots.BotID) (dto.ThreadsTable, error)
}
