package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func operationFromDTO(d string) (bots.Operation, error) {
	switch d {
	case "noop":
		return bots.NoOp{}, nil
	case "save":
		return bots.SaveOp{}, nil
	case "append":
		return bots.AppendOp{}, nil
	default:
		return nil, bots.NewValidationError(bots.NewValidationErrorDetail(
			"", "operation-invalid-type",
			fmt.Sprintf("expected operation one of ['noop', 'save', 'append'], got '%s'", d),
		))
	}
}

func operationsFromDTOPrefixed(ds []string, prefix string) ([]bots.Operation, error) {
	var vErr bots.ValidationError

	ops := make([]bots.Operation, len(ds))
	for i, d := range ds {
		op, err := operationFromDTO(d)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		ops[i] = op
	}

	if vErr.OrError() != nil {
		return nil, vErr
	}
	return ops, nil
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
