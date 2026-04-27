package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type OwnedMailingProvider interface {
	OwnedMailing(ctx context.Context, id bots.MailingID) (dto.OwnedMailing, error)
}
