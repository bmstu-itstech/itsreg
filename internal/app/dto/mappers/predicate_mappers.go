package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func PredicateFromDTO(d dto.Predicate) (bots.Predicate, error) {
	switch d.Type {
	case "always":
		return bots.AlwaysTruePredicate{}, nil
	case "exact":
		return bots.NewExactMatchPredicate(d.Data)
	case "regex":
		return bots.NewRegexMatchPredicate(d.Data)
	default:
		return nil, bots.NewValidationError(bots.NewValidationErrorDetail(
			"type", "predicate-invalid-type",
			fmt.Sprintf("expected predicate type one of ['always', 'exact', 'regex'], got '%s'", d.Type),
		))
	}
}

func PredicateToDTO(p bots.Predicate) dto.Predicate {
	switch pt := p.(type) {
	case bots.AlwaysTruePredicate:
		return dto.Predicate{Type: "always", Data: ""}

	case bots.ExactMatchPredicate:
		return dto.Predicate{Type: "exact", Data: pt.Text()}

	case bots.RegexMatchPredicate:
		return dto.Predicate{Type: "regex", Data: pt.Pattern()}

	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}
