package bots

import "github.com/bmstu-itstech/itsreg/internal/domain/shared"

const ErrorCodeStateInvalid = "state-invalid"

// State есть состояние в контексте FSM и уникальный номер узла в пределах скрипта.
type State struct {
	i int
}

var ZeroState State

func NewState(i int) (State, error) {
	if i <= 0 {
		return ZeroState, shared.NewValidationError(shared.NewValidationErrorDetail(
			"value",
			ErrorCodeStateInvalid,
			"state cannot be less or equal than zero",
		))
	}

	return State{i: i}, nil
}

func MustNewState(i int) State {
	s, err := NewState(i)
	if err != nil {
		panic(err)
	}
	return s
}

func (s State) Int() int {
	return s.i
}
