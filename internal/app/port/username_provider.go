package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var ErrUsernameNotFound = errors.New("username not found")

type UsernameProvider interface {
	// Username возвращает имя пользователя с UserID.
	Username(ctx context.Context, id bots.UserID) (bots.Username, error)
}
