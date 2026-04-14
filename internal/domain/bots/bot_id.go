package bots

import "github.com/bmstu-itstech/itsreg/pkg/nanoid"

const BotIDLen = 6

// BotID есть уникальный идентификатор бота.
type BotID string

func NewBotID() BotID {
	return BotID(nanoid.NewNanoID(BotIDLen))
}

func (id BotID) IsZero() bool {
	return id == ""
}

func (id BotID) String() string {
	return string(id)
}
