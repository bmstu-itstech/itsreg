package bots

type EntryKey string

type Entry struct {
	key   EntryKey
	start State
}

func NewEntry(key EntryKey, start State) (Entry, error) {
	if start < 0 {
		return Entry{}, NewInvalidInputError(
			"invalid-entry",
			"expected non-negative start state",
		)
	}

	if key == "" {
		return Entry{}, NewInvalidInputError(
			"invalid-entry-key",
			"failed to create entry: expected not empty key",
		)
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
