package bots

import "fmt"

// Operation описывает действие, которое будет произведено над Participant
// после обработки Message.
type Operation interface {
	Apply(thr *Thread, in Message)
}

// NewOperationFromString создаёт Operation из строк:
// - "noop" - NoOp,
// - "save" - SaveOp
// - "append" - AppendOp.
func NewOperationFromString(s string) (Operation, error) {
	switch s {
	case "noop":
		return NoOp{}, nil
	case "save":
		return SaveOp{}, nil
	case "append":
		return AppendOp{}, nil
	}
	return nil, NewInvalidInputError(
		"invalid-operation",
		fmt.Sprintf("failed to create operaion: expected one of ['noop', 'save', 'append'], got %s", s),
	)
}

func MustNewOperationFromString(s string) Operation {
	a, err := NewOperationFromString(s)
	if err != nil {
		panic(err)
	}
	return a
}

// NoOp не производит никаких действий над пользователем.
type NoOp struct{}

func (a NoOp) Apply(_ *Thread, _ Message) {}

// SaveOp вызывает для пользователя Participant.SaveAnswer.
type SaveOp struct{}

func (a SaveOp) Apply(thr *Thread, in Message) {
	thr.SaveAnswer(in)
}

// AppendOp вызывает для пользователя Participant.AppendAnswer.
type AppendOp struct{}

func (a AppendOp) Apply(thr *Thread, in Message) {
	thr.AppendAnswer(in)
}
