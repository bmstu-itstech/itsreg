package dto

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Predicate struct {
	Type string
	Data string // Любым образом сериализованные данные о предикате, зависит от Type
}

type predicateBuilder struct {
	state int
}

func (b *predicateBuilder) WithState(state int) *predicateBuilder {
	b.state = state
	return b
}

func (b *predicateBuilder) Build(dto Predicate) (bots.Predicate, error) {
	p, err := b.fromDTO(dto)
	if err != nil {
		return nil, b.enrichError(err)
	}
	return p, nil
}

func (b *predicateBuilder) fromDTO(dto Predicate) (bots.Predicate, error) {
	switch dto.Type {
	case "always":
		return bots.AlwaysTruePredicate{}, nil

	case "exact":
		return bots.NewExactMatchPredicate(dto.Data)

	case "regex":
		return bots.NewRegexMatchPredicate(dto.Data)

	default:
		return nil, bots.NewInvalidInputError(
			"predicate-invalid-type",
			fmt.Sprintf("expected predicate type one of ['always', 'exact', 'regex'], got '%s'", dto.Type),
			"field",
			"type",
		)
	}
}

func (b *predicateBuilder) enrichError(err error) error {
	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		if b.state != 0 {
			iiErr.Details["state"] = strconv.Itoa(b.state)
		}
		return iiErr
	}
	return err
}

func predicateToDTO(p bots.Predicate) Predicate {
	switch p := p.(type) {
	case bots.AlwaysTruePredicate:
		return Predicate{Type: "always", Data: ""}

	case bots.ExactMatchPredicate:
		return Predicate{Type: "exact", Data: p.Text()}

	case bots.RegexMatchPredicate:
		return Predicate{Type: "regex", Data: p.Pattern()}

	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}
