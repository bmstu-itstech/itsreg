package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type BotMetaProvider interface {
	BotMeta(ctx context.Context, id bots.BotID) (dto.BotMeta, error)
}
