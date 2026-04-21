package apiv3

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
)

func predicateFromAPI(p Predicate) dto.Predicate {
	d, err := p.Discriminator()
	if err != nil {
		return dto.Predicate{}
	}

	switch d {
	case string(Always):
		return dto.Predicate{
			Type: string(Always),
		}

	case string(Exact):
		exact, err2 := p.AsExactPredicate()
		if err2 != nil {
			return dto.Predicate{}
		}
		return dto.Predicate{
			Type: string(Exact),
			Data: exact.Text,
		}

	case string(Regex):
		regexp, err2 := p.AsRegexPredicate()
		if err2 != nil {
			return dto.Predicate{}
		}
		return dto.Predicate{
			Type: string(Regex),
			Data: regexp.Pattern,
		}

	default:
		return dto.Predicate{
			Type: d,
		}
	}
}

func edgeFromAPI(e Edge) dto.Edge {
	return dto.Edge{
		Predicate: predicateFromAPI(e.Predicate),
		To:        e.To,
		Operation: string(e.Operation),
	}
}

func edgesFromAPI(es []Edge) []dto.Edge {
	edges := make([]dto.Edge, len(es))
	for i, e := range es {
		edges[i] = edgeFromAPI(e)
	}
	return edges
}

func messageFromAPI(m Message) dto.Message {
	return dto.Message{
		Text: m.Text,
	}
}

func messagesFromAPI(ms []Message) []dto.Message {
	messages := make([]dto.Message, len(ms))
	for i, m := range ms {
		messages[i] = messageFromAPI(m)
	}
	return messages
}

func nodeFromAPI(n Node) dto.Node {
	return dto.Node{
		State:    n.State,
		Title:    n.Title,
		Edges:    edgesFromAPI(derefOrNilSlice(n.Edges)),
		Messages: messagesFromAPI(n.Messages),
		Options:  derefOrNilSlice(n.Options),
	}
}

func nodesFromAPI(ns []Node) []dto.Node {
	nodes := make([]dto.Node, len(ns))
	for i, n := range ns {
		nodes[i] = nodeFromAPI(n)
	}
	return nodes
}

func entryFromAPI(e Entry) dto.Entry {
	return dto.Entry{
		Key:   e.Key,
		Start: e.Start,
	}
}

func entriesFromAPI(es []Entry) []dto.Entry {
	entries := make([]dto.Entry, len(es))
	for i, e := range es {
		entries[i] = entryFromAPI(e)
	}
	return entries
}
