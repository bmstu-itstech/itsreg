package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
)

type InboundDispatcher interface {
	Dispatch(ctx context.Context, in dto.InboundMessage) error
}
