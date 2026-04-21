package bots

import "github.com/bmstu-itstech/itsreg/pkg/nanoid"

const ThreadIDLen = 8

type ThreadID string

func NewThreadID() ThreadID {
	return ThreadID(nanoid.NewNanoID(ThreadIDLen))
}

func (t ThreadID) IsZero() bool {
	return t == ""
}

func (t ThreadID) String() string {
	return string(t)
}
