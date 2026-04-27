package bots

import "github.com/bmstu-itstech/itsreg/pkg/nanoid"

const MailingIDLen = 6

// MailingID есть уникальный идентификатор рассылки.
type MailingID string

func NewMailingID() MailingID {
	return MailingID(nanoid.NewNanoID(MailingIDLen))
}

func (id MailingID) IsZero() bool {
	return id == ""
}

func (id MailingID) String() string {
	return string(id)
}
