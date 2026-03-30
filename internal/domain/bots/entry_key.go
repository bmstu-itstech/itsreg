package bots

// EntryKey есть уникальный идентификатор Entry в пределах одного Script.
type EntryKey string

func (e EntryKey) IsZero() bool {
	return e == ""
}

func (e EntryKey) String() string {
	return string(e)
}
