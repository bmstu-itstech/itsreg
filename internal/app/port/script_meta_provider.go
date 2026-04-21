package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type ScriptMetaProvider interface {
	ScriptMeta(ctx context.Context, id bots.ScriptID) (dto.ScriptMeta, error)
}
