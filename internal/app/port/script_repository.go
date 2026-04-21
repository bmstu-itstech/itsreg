package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var (
	ErrScriptNotFound      = errors.New("script not found")
	ErrScriptAlreadyExists = errors.New("script already exists")
)

type ScriptRepository interface {
	Script(ctx context.Context, id bots.ScriptID) (*bots.Script, error)
	ScriptsByOwnerID(ctx context.Context, ownerID bots.UserID) ([]*bots.Script, error)

	SaveScript(ctx context.Context, s *bots.Script) error
	UpdateScript(ctx context.Context, s *bots.Script) error
}
