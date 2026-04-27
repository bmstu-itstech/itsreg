package port

import (
	"context"
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

var (
	ErrMailingAlreadyExists = errors.New("mailing already exists")
	ErrMailingNotFound      = errors.New("mailing not found")
)

type MailingsFilter struct {
	BotID  *bots.BotID
	Status *bots.MailingStatus
}

type MailingRepository interface {
	Mailing(ctx context.Context, id bots.MailingID) (*bots.Mailing, error)
	MailingsByOwnerID(ctx context.Context, ownerID bots.UserID, filter MailingsFilter) ([]*bots.Mailing, error)

	SaveMailing(ctx context.Context, mailing *bots.Mailing) error
	UpdateMailing(ctx context.Context, mailing *bots.Mailing) error
}
