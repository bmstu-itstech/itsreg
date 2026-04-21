package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var ErrTokenInvalid = errors.New("token invalid")

type TokenService interface {
	VerifyToken(ctx context.Context, token string) (bots.UserID, error)
}
