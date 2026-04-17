package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

func nodeFromDTO(d dto.Node) (bots.Node, error) {
	var vErr shared.ValidationError

	state, err := bots.NewState(d.State)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "state")
	}

	edges, err := edgesFromDTOPrefixed(d.Edges, "edges")
	if err != nil {
		vErr = vErr.AppendError(err)
	}

	messages, err := messagesFromDTOPrefixed(d.Messages, "messages")
	if err != nil {
		vErr = vErr.AppendError(err)
	}

	options, err := optionsFromDTOPrefixed(d.Options, "options")
	if err != nil {
		vErr = vErr.AppendError(err)
	}

	if vErr.OrError() != nil {
		return bots.Node{}, vErr
	}

	return bots.NewNode(state, d.Title, edges, messages, options)
}

func NodesFromDTOPrefixed(ds []dto.Node, prefix string) ([]bots.Node, error) {
	var vErr shared.ValidationError

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
