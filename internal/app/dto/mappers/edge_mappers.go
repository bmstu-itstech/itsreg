package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

func edgeFromDTO(d dto.Edge) (bots.Edge, error) {
	var vErr shared.ValidationError

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

func edgesFromDTOPrefixed(ds []dto.Edge, prefix string) ([]bots.Edge, error) {
	var vErr shared.ValidationError
	edges := make([]bots.Edge, len(ds))
	for i, de := range ds {
		edge, err2 := edgeFromDTO(de)
		if err2 != nil {
			vErr = vErr.AppendPrefixed(err2, fmt.Sprintf("%s[%d]", prefix, i))
		}
		edges[i] = edge
	}
	return edges, vErr
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
