package dto

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type operationBuilder struct {
	state int
}

func (b *operationBuilder) WithState(state int) *operationBuilder {
	b.state = state
	return b
}

func (b *operationBuilder) Build(dto string) (bots.Operation, error) {
	o, err := b.fromDTO(dto)
	if err != nil {
		return nil, b.enrichError(err)
	}
	return o, nil
}

func (b *operationBuilder) fromDTO(dto string) (bots.Operation, error) {
	switch dto {
	case "noop":
		return bots.NoOp{}, nil
	case "save":
		return bots.SaveOp{}, nil
	case "append":
		return bots.AppendOp{}, nil
	default:
		return nil, bots.NewInvalidInputError(
			"operation-invalid-type",
			fmt.Sprintf("expected operation type one of ['noop', 'save', 'append'], got '%s'", dto),
		)
	}
}

func (b *operationBuilder) enrichError(err error) error {
	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		if b.state != 0 {
			iiErr.Details["state"] = strconv.Itoa(b.state)
		}
		return iiErr
	}
	return err
}

func operationToDTO(op bots.Operation) string {
	switch op.(type) {
	case bots.NoOp:
		return "noop"
	case bots.SaveOp:
		return "save"
	case bots.AppendOp:
		return "append"
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}
