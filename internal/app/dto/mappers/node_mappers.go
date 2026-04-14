package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func nodeFromDTO(d dto.Node) (bots.Node, error) {
	var vErr bots.ValidationError

	state, err := bots.NewState(d.State)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "state")
	}

	edges := make([]bots.Edge, len(d.Edges))
	for i, de := range d.Edges {
		edge, err2 := edgeFromDTO(de)
		if err2 != nil {
			vErr = vErr.AppendPrefixed(err2, fmt.Sprintf("edges[%d]", i))
		}
		edges[i] = edge
	}

	messages := make([]bots.Message, len(d.Messages))
	for i, dm := range d.Messages {
		message, err2 := MessageFromDTO(dm)
		if err2 != nil {
			vErr = vErr.AppendPrefixed(err2, fmt.Sprintf("messages[%d]", i))
		}
		messages[i] = message
	}

	options := make([]bots.Option, len(d.Options))
	for i, do := range d.Options {
		option, err2 := bots.NewOption(do)
		if err2 != nil {
			vErr = vErr.AppendPrefixed(err2, fmt.Sprintf("options[%d]", i))
		}
		options[i] = option
	}

	if vErr.OrError() != nil {
		return bots.Node{}, vErr
	}

	return bots.NewNode(state, d.Title, edges, messages, options)
}

func NodesFromDTOPrefixed(ds []dto.Node, prefix string) ([]bots.Node, error) {
	var vErr bots.ValidationError

	nodes := make([]bots.Node, len(ds))
	for i, n := range ds {
		node, err := nodeFromDTO(n)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		nodes[i] = node
	}

	if vErr.OrError() != nil {
		return nil, vErr
	}
	return nodes, nil
}

func nodeToDTO(n bots.Node) dto.Node {
	return dto.Node{
		State:    n.State().Int(),
		Title:    n.Title(),
		Edges:    edgesToDTO(n.Edges()),
		Messages: messagesToDTO(n.Messages()),
		Options:  optionsToDTO(n.Options()),
	}
}

func nodesToDTO(ns []bots.Node) []dto.Node {
	res := make([]dto.Node, 0, len(ns))
	for _, n := range ns {
		res = append(res, nodeToDTO(n))
	}
	return res
}
