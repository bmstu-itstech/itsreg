package bots

import "github.com/bmstu-itstech/itsreg/pkg/nanoid"

const ScriptIDLen = 6

type ScriptID string

func NewScriptID() ScriptID {
	return ScriptID(nanoid.NewNanoID(ScriptIDLen))
}

func (id ScriptID) String() string {
	return string(id)
}

func (id ScriptID) IsZero() bool {
	return id == ""
}
