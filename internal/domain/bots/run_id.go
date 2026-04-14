package bots

import "github.com/bmstu-itstech/itsreg/pkg/nanoid"

const RunIDLen = 6

type RunID string

func NewRunID() RunID {
	return RunID(nanoid.NewNanoID(RunIDLen))
}

func (id RunID) String() string {
	return string(id)
}

func (id RunID) IsZero() bool {
	return id == ""
}
