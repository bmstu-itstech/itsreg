package bots

import (
	"errors"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

const ErrorCodeEntryEmptyKey = "entry-empty-key"

type Entry struct {
	key   EntryKey
	start State
}

func NewEntry(key EntryKey, start State) (Entry, error) {
	if key == "" {
		return Entry{}, shared.NewValidationError(
			shared.NewValidationErrorDetail("key", ErrorCodeEntryEmptyKey, "entry key cannot be empty"),
		)
	}

	if start == ZeroState {
		return Entry{}, errors.New("empty start state")
	}

	return Entry{
		key:   key,
		start: start,
	}, nil
}

func MustNewEntry(key EntryKey, start State) Entry {
	e, err := NewEntry(key, start)
	if err != nil {
		panic(err)
	}
	return e
}

func (e Entry) IsZero() bool {
	return e == Entry{}
}

func (e Entry) Key() EntryKey {
	return e.key
}

func (e Entry) Start() State {
	return e.start
}
