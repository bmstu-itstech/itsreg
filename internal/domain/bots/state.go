package bots

import "fmt"

// State есть состояние в контексте FSM и уникальный номер узла в пределах скрипта.
type State struct {
	i int
}

var ZeroState State

func NewState(i int) (State, error) {
	if i == 0 {
		return State{}, NewInvalidInputError("state-empty", "expected not empty state")
	}
	if i < 0 {
		return ZeroState, NewInvalidInputError(
			"state-invalid",
			fmt.Sprintf("expected state is positive value, got %d", i),
		)
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
