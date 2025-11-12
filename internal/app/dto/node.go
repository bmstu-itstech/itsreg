package dto

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Node struct {
	State    int
	Title    string
	Edges    []Edge
	Messages []Message
	Options  []string
}

func nodeFromDTO(dto Node) (bots.Node, error) {
	var errs bots.MultiError

	state, err := bots.NewState(dto.State)
	if err != nil {
		return bots.Node{}, err
	}

	eb := new(edgeBuilder).WithState(dto.State)
	mb := new(messageBuilder).WithState(dto.State)
	ob := new(optionBuilder).WithState(dto.State)

	es, err := eb.BuildAll(dto.Edges)
	if err != nil {
		errs.ExtendOrAppend(err)
	}

	ms, err := mb.BuildAll(dto.Messages)
	if err != nil {
		errs.ExtendOrAppend(err)
	}

	os, err := ob.BuildAll(dto.Options)
	if err != nil {
		errs.ExtendOrAppend(err)
	}

	if errs.HasError() {
		return bots.Node{}, &errs
	}

	return bots.NewNode(state, dto.Title, es, ms, os)
}

func batchNodesFromDTO(dtos []Node) ([]bots.Node, error) {
	var errs bots.MultiError
	res := make([]bots.Node, len(dtos))
	for i, dto := range dtos {
		n, err := nodeFromDTO(dto)
		if err != nil {
			errs.ExtendOrAppend(err)
		}
		res[i] = n
	}
	if errs.HasError() {
		return nil, &errs
	}
	return res, nil
}

func nodeToDTO(node bots.Node) Node {
	return Node{
		State:    node.State().Int(),
		Title:    node.Title(),
		Edges:    batchEdgesToDTO(node.Edges()),
		Messages: batchMessagesToDTO(node.Messages()),
		Options:  batchOptionsToDTO(node.Options()),
	}
}

func batchNodesToDto(nodes []bots.Node) []Node {
	res := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		res = append(res, nodeToDTO(node))
	}
	return res
}
