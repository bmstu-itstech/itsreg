package bots

type EntryKey string

type Entry struct {
	key   EntryKey
	start State
}

func NewEntry(key EntryKey, start State) (Entry, error) {
	if key == "" {
		return Entry{}, NewInvalidInputError(
			"invalid-entry-key",
			"failed to create entry: expected not empty key",
		)
	}

	if start == ZeroState {
		return Entry{}, NewInvalidInputError(
			"invalid-entry-start",
			"failed to create entry: invalid start state",
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
