package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func edgeFromDTO(d dto.Edge) (bots.Edge, error) {
	var vErr bots.ValidationError

	pred, err := PredicateFromDTO(d.Predicate)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "predicate")
	}

	oper, err := operationFromDTO(d.Operation)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "operation")
	}

	to, err := bots.NewState(d.To)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "to")
	}

	if vErr.OrError() != nil {
		return bots.Edge{}, vErr
	}
	return bots.NewEdge(pred, to, oper), nil
}

func edgesFromDTOPrefixed(prefix string, ds []dto.Edge) ([]bots.Edge, error) {
	var vErr bots.ValidationError

	edges := make([]bots.Edge, len(ds))
	for i, d := range ds {
		edge, err := edgeFromDTO(d)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		edges[i] = edge
	}

	if vErr.OrError() != nil {
		return nil, vErr
	}
	return edges, nil
}

func edgeToDTO(e bots.Edge) dto.Edge {
	return dto.Edge{
		Predicate: PredicateToDTO(e.Predicate),
		To:        e.To().Int(),
		Operation: operationToDTO(e.Operation()),
	}
}

func edgesToDTO(es []bots.Edge) []dto.Edge {
	res := make([]dto.Edge, 0, len(es))
	for _, edge := range es {
		res = append(res, edgeToDTO(edge))
	}
	return res
}
