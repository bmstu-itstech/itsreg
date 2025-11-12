package http

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
)

func batchBotsFromApp(bots []dto.Bot) []Bot {
	res := make([]Bot, len(bots))
	for i, bot := range bots {
		res[i] = botFromApp(bot)
	}
	return res
}

func botFromApp(bot dto.Bot) Bot {
	return Bot{
		Author: bot.Author,
		Id:     bot.ID,
		Script: scriptFromApp(bot.Script),
		Token:  bot.Token,
	}
}

func scriptToApp(bot Script) (dto.Script, error) {
	nodes, err := batchNodeToApp(bot.Nodes)
	if err != nil {
		return dto.Script{}, err
	}

	return dto.Script{
		Nodes:   nodes,
		Entries: batchEntriesToApp(bot.Entries),
	}, nil
}

func scriptFromApp(bot dto.Script) Script {
	return Script{
		Entries: batchEntriesFromApp(bot.Entries),
		Nodes:   batchNodesFromApp(bot.Nodes),
	}
}

func entryToApp(entry Entry) dto.Entry {
	return dto.Entry{
		Key:   entry.Key,
		Start: entry.Start,
	}
}

func entryFromApp(entry dto.Entry) Entry {
	return Entry{
		Key:   entry.Key,
		Start: entry.Start,
	}
}

func batchEntriesToApp(entries []Entry) []dto.Entry {
	res := make([]dto.Entry, len(entries))
	for i, entry := range entries {
		res[i] = entryToApp(entry)
	}
	return res
}

func batchEntriesFromApp(entries []dto.Entry) []Entry {
	res := make([]Entry, len(entries))
	for i, entry := range entries {
		res[i] = entryFromApp(entry)
	}
	return res
}

func nodeToApp(node Node) (dto.Node, error) {
	edges, err := batchEdgesToApp(emptyOnNil(node.Edges))
	if err != nil {
		return dto.Node{}, err
	}

	return dto.Node{
		State:    node.State,
		Title:    node.Title,
		Edges:    edges,
		Messages: batchMessageToApp(node.Messages),
		Options:  emptyOnNil(node.Options),
	}, nil
}

func nodeFromApp(node dto.Node) Node {
	return Node{
		Edges:    nilOnEmpty(batchEdgesFromApp(node.Edges)),
		Messages: batchMessagesFromApp(node.Messages),
		State:    node.State,
		Title:    node.Title,
		Options:  nilOnEmpty(node.Options),
	}
}

func batchNodeToApp(nodes []Node) ([]dto.Node, error) {
	res := make([]dto.Node, len(nodes))
	for i, node := range nodes {
		n, err := nodeToApp(node)
		if err != nil {
			return nil, err
		}
		res[i] = n
	}
	return res, nil
}

func batchNodesFromApp(nodes []dto.Node) []Node {
	res := make([]Node, len(nodes))
	for i, node := range nodes {
		res[i] = nodeFromApp(node)
	}
	return res
}

func edgeToApp(edge Edge) (dto.Edge, error) {
	pred, err := predicateToApp(edge.Predicate)
	if err != nil {
		return dto.Edge{}, err
	}

	return dto.Edge{
		Predicate: pred,
		To:        edge.To,
		Operation: string(edge.Operation),
	}, nil
}

func edgeFromApp(edge dto.Edge) Edge {
	return Edge{
		Operation: EdgeOperation(edge.Operation),
		Predicate: predicateFromApp(edge.Predicate),
		To:        edge.To,
	}
}

func batchEdgesToApp(edges []Edge) ([]dto.Edge, error) {
	res := make([]dto.Edge, len(edges))
	for i, edge := range edges {
		e, err := edgeToApp(edge)
		if err != nil {
			return nil, err
		}
		res[i] = e
	}
	return res, nil
}

func batchEdgesFromApp(edges []dto.Edge) []Edge {
	res := make([]Edge, len(edges))
	for i, edge := range edges {
		res[i] = edgeFromApp(edge)
	}
	return res
}

func batchMessageToApp(msgs []Message) []dto.Message {
	res := make([]dto.Message, len(msgs))
	for i, msg := range msgs {
		res[i] = messageToApp(msg)
	}
	return res
}

func batchMessagesFromApp(msgs []dto.Message) []Message {
	res := make([]Message, len(msgs))
	for i, msg := range msgs {
		res[i] = messageFromApp(msg)
	}
	return res
}

func predicateToApp(pred Predicate) (dto.Predicate, error) {
	d, err := pred.Discriminator()
	if err != nil {
		return dto.Predicate{}, err
	}

	switch d {
	case string(Always):
		return dto.Predicate{
			Type: string(Always),
		}, nil

	case string(Exact):
		exact, err2 := pred.AsExactPredicate()
		if err2 != nil {
			return dto.Predicate{}, err2
		}
		return dto.Predicate{
			Type: string(Exact),
			Data: exact.Text,
		}, err2

	case string(Regex):
		regexp, err2 := pred.AsRegexPredicate()
		if err2 != nil {
			return dto.Predicate{}, err2
		}
		return dto.Predicate{
			Type: string(Regex),
			Data: regexp.Pattern,
		}, err2

	default:
		return dto.Predicate{}, fmt.Errorf(
			"invalid predicate type %s, expected one of ['always', 'exact', 'regexp']", d,
		)
	}
}

func predicateFromApp(pred dto.Predicate) Predicate {
	switch pred.Type {
	case string(Always):
		p := Predicate{}
		_ = p.FromAlwaysPredicate(AlwaysPredicate{Type: Always})
		return p

	case string(Exact):
		p := Predicate{}
		_ = p.FromExactPredicate(ExactPredicate{
			Type: Exact,
			Text: pred.Data,
		})
		return p

	case string(Regex):
		p := Predicate{}
		_ = p.FromRegexPredicate(RegexPredicate{
			Type:    Regex,
			Pattern: pred.Data,
		})
		return p

	default:
		return Predicate{}
	}
}

func messageToApp(message Message) dto.Message {
	return dto.Message{
		Text: message.Text,
	}
}

func messageFromApp(message dto.Message) Message {
	return Message{
		Text: message.Text,
	}
}

func emptyOnNil[T any](ts *[]T) []T {
	if ts == nil {
		return []T{}
	}
	return *ts
}

func nilOnEmpty[T any](ts []T) *[]T {
	if len(ts) == 0 {
		return nil
	}
	return &ts
}

func nilOnEmptyMap[K comparable, V any](tm map[K]V) *map[K]V {
	if len(tm) == 0 {
		return nil
	}
	return &tm
}
