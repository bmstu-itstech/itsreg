package dto

import "github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"

type Script struct {
	Nodes   []Node
	Entries []Entry
}

func ScriptFromDTO(dto Script) (bots.Script, error) {
	nodes, err := batchNodesFromDTO(dto.Nodes)
	if err != nil {
		return bots.Script{}, err
	}

	entries, err := batchEntriesFromDto(dto.Entries)
	if err != nil {
		return bots.Script{}, err
	}

	return bots.NewScript(nodes, entries)
}

func scriptToDTO(script bots.Script) Script {
	return Script{
		Nodes:   batchNodesToDto(script.Nodes()),
		Entries: batchEntriesToDto(script.Entries()),
	}
}
