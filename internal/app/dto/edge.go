package dto

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Edge struct {
	Predicate Predicate
	To        int
	Operation string
}
type edgeBuilder struct {
	p predicateBuilder
	o operationBuilder
}

func (b *edgeBuilder) WithState(state int) *edgeBuilder {
	b.p.WithState(state)
	b.o.WithState(state)
	return b
}

func (b *edgeBuilder) Build(dto Edge) (bots.Edge, error) {
	var errs bots.MultiError

	p, err := b.p.Build(dto.Predicate)
	if err != nil {
		errs.Append(err)
	}

	o, err := b.o.Build(dto.Operation)
	if err != nil {
		errs.Append(err)
	}

	if errs.HasError() {
		return bots.Edge{}, &errs
	}

	to, err := bots.NewState(dto.To)
	if err != nil {
		return bots.Edge{}, err
	}

	return bots.NewEdge(p, to, o), nil
}

func (b *edgeBuilder) BuildAll(dtos []Edge) ([]bots.Edge, error) {
	var errs bots.MultiError
	res := make([]bots.Edge, len(dtos))
	for i, dto := range dtos {
		e, err := b.Build(dto)
		if err != nil {
			errs.ExtendOrAppend(err)
		}
		res[i] = e
	}
	if errs.HasError() {
		return res, &errs
	}
	return res, nil
}

func edgeToDTO(e bots.Edge) Edge {
	return Edge{
		Predicate: predicateToDTO(e.Predicate),
		To:        e.To().Int(),
		Operation: operationToDTO(e.Operation()),
	}
}

func batchEdgesToDTO(edges []bots.Edge) []Edge {
	res := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		res = append(res, edgeToDTO(edge))
	}
	return res
}
