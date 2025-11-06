package http

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg-bots/internal/app"
)

func batchBotsFromApp(bots []app.Bot) []Bot {
	res := make([]Bot, len(bots))
	for i, bot := range bots {
		res[i] = botFromApp(bot)
	}
	return res
}

func botFromApp(bot app.Bot) Bot {
	return Bot{
		Author: bot.Author,
		Id:     bot.Id,
		Script: scriptFromApp(bot.Script),
		Token:  bot.Token,
	}
}

func scriptToApp(bot Script) (app.Script, error) {
	nodes, err := batchNodeToApp(bot.Nodes)
	if err != nil {
		return app.Script{}, err
	}

	return app.Script{
		Nodes:   nodes,
		Entries: batchEntriesToApp(bot.Entries),
	}, nil
}

func scriptFromApp(bot app.Script) Script {
	return Script{
		Entries: batchEntriesFromApp(bot.Entries),
		Nodes:   batchNodesFromApp(bot.Nodes),
	}
}

func entryToApp(entry Entry) app.Entry {
	return app.Entry{
		Key:   entry.Key,
		Start: uint(entry.Start),
	}
}

func entryFromApp(entry app.Entry) Entry {
	return Entry{
		Key:   entry.Key,
		Start: int(entry.Start),
	}
}

func batchEntriesToApp(entries []Entry) []app.Entry {
	res := make([]app.Entry, len(entries))
	for i, entry := range entries {
		res[i] = entryToApp(entry)
	}
	return res
}

func batchEntriesFromApp(entries []app.Entry) []Entry {
	res := make([]Entry, len(entries))
	for i, entry := range entries {
		res[i] = entryFromApp(entry)
	}
	return res
}

func nodeToApp(node Node) (app.Node, error) {
	edges, err := batchEdgesToApp(emptyOnNil(node.Edges))
	if err != nil {
		return app.Node{}, err
	}

	return app.Node{
		State:    uint(node.State),
		Title:    node.Title,
		Edges:    edges,
		Messages: batchMessageToApp(node.Messages),
		Options:  emptyOnNil(node.Options),
	}, nil
}

func nodeFromApp(node app.Node) Node {
	return Node{
		Edges:    nilOnEmpty(batchEdgesFromApp(node.Edges)),
		Messages: batchMessagesFromApp(node.Messages),
		State:    int(node.State),
		Title:    node.Title,
		Options:  nilOnEmpty(node.Options),
	}
}

func batchNodeToApp(nodes []Node) ([]app.Node, error) {
	res := make([]app.Node, len(nodes))
	for i, node := range nodes {
		n, err := nodeToApp(node)
		if err != nil {
			return nil, err
		}
		res[i] = n
	}
	return res, nil
}

func batchNodesFromApp(nodes []app.Node) []Node {
	res := make([]Node, len(nodes))
	for i, node := range nodes {
		res[i] = nodeFromApp(node)
	}
	return res
}

func edgeToApp(edge Edge) (app.Edge, error) {
	pred, err := predicateToApp(edge.Predicate)
	if err != nil {
		return app.Edge{}, err
	}

	return app.Edge{
		Predicate: pred,
		To:        uint(edge.To),
		Operation: string(edge.Operation),
	}, nil
}

func edgeFromApp(edge app.Edge) Edge {
	return Edge{
		Operation: EdgeOperation(edge.Operation),
		Predicate: predicateFromApp(edge.Predicate),
		To:        int(edge.To),
	}
}

func batchEdgesToApp(edges []Edge) ([]app.Edge, error) {
	res := make([]app.Edge, len(edges))
	for i, edge := range edges {
		e, err := edgeToApp(edge)
		if err != nil {
			return nil, err
		}
		res[i] = e
	}
	return res, nil
}

func batchEdgesFromApp(edges []app.Edge) []Edge {
	res := make([]Edge, len(edges))
	for i, edge := range edges {
		res[i] = edgeFromApp(edge)
	}
	return res
}

func batchMessageToApp(msgs []Message) []app.Message {
	res := make([]app.Message, len(msgs))
	for i, msg := range msgs {
		res[i] = messageToApp(msg)
	}
	return res
}

func batchMessagesFromApp(msgs []app.Message) []Message {
	res := make([]Message, len(msgs))
	for i, msg := range msgs {
		res[i] = messageFromApp(msg)
	}
	return res
}

func predicateToApp(pred Predicate) (app.Predicate, error) {
	d, err := pred.Discriminator()
	if err != nil {
		return app.Predicate{}, err
	}

	switch d {
	case string(Always):
		return app.Predicate{
			Type: string(Always),
		}, nil

	case string(Exact):
		exact, err := pred.AsExactPredicate()
		if err != nil {
			return app.Predicate{}, err
		}
		return app.Predicate{
			Type: string(Exact),
			Data: exact.Text,
		}, err

	case string(Regexp):
		regexp, err := pred.AsRegexpPredicate()
		if err != nil {
			return app.Predicate{}, err
		}
		return app.Predicate{
			Type: string(Regexp),
			Data: regexp.Pattern,
		}, err

	default:
		return app.Predicate{}, fmt.Errorf("invalid predicate type %s, expected one of ['always', 'exact', 'regexp']", d)
	}
}

func predicateFromApp(pred app.Predicate) Predicate {
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

	case string(Regexp):
		p := Predicate{}
		_ = p.FromRegexpPredicate(RegexpPredicate{
			Type:    Regexp,
			Pattern: pred.Data,
		})
		return p

	default:
		return Predicate{}
	}
}

func messageToApp(message Message) app.Message {
	return app.Message{
		Text: message.Text,
	}
}

func messageFromApp(message app.Message) Message {
	return Message{
		Text: message.Text,
	}
}

func emptyOnNil[T any](ts *[]T) []T {
	if ts == nil {
		return []T{}
	} else {
		return *ts
	}
}

func nilOnEmpty[T any](ts []T) *[]T {
	if ts == nil {
		return nil
	} else {
		return &ts
	}
}
