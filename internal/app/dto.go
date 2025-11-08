package app

import (
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Predicate struct {
	Type string
	Data string // Любым образом сериализованные данные о предикате, зависит от Type
}

type Edge struct {
	Predicate Predicate
	To        int
	Operation string
}

type Message struct {
	Text string
}

type Node struct {
	State    int
	Title    string
	Edges    []Edge
	Messages []Message
	Options  []string
}

type Entry struct {
	Key   string
	Start int
}

type Script struct {
	Nodes   []Node
	Entries []Entry
}

type Bot struct {
	ID     string
	Token  string
	Author int64
	Script Script
}

type Thread struct {
	ID        string
	Key       string
	StartedAt time.Time
	Username  string
	Answers   map[int]Message
}

func predicateFromDto(dto Predicate) (bots.Predicate, error) {
	switch dto.Type {
	case "always":
		return bots.AlwaysTruePredicate{}, nil

	case "exact":
		return bots.NewExactMatchPredicate(dto.Data)

	case "regexp":
		return bots.NewRegexMatchPredicate(dto.Data)

	default:
		return nil, bots.NewInvalidInputError(
			"invalid-predicate",
			fmt.Sprintf("expected predicate type one of ['always', 'exact', 'regexp'], got %s", dto.Type),
		)
	}
}

func predicateToDto(p bots.Predicate) Predicate {
	switch p := p.(type) {
	case bots.AlwaysTruePredicate:
		return Predicate{Type: "always", Data: ""}

	case bots.ExactMatchPredicate:
		return Predicate{Type: "exact", Data: p.Text()}

	case bots.RegexMatchPredicate:
		return Predicate{Type: "regexp", Data: p.Pattern()}

	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func operationFromDto(dto string) (bots.Operation, error) {
	switch dto {
	case "noop":
		return bots.NoOp{}, nil
	case "save":
		return bots.SaveOp{}, nil
	case "append":
		return bots.AppendOp{}, nil
	default:
		return nil, bots.NewInvalidInputError(
			"invalid-operation",
			fmt.Sprintf("expected operation type one of ['noop', 'save', 'append'], got %s", dto),
		)
	}
}

func operationToDto(op bots.Operation) string {
	switch op.(type) {
	case bots.NoOp:
		return "noop"
	case bots.SaveOp:
		return "save"
	case bots.AppendOp:
		return "append"
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func edgeFromDto(dto Edge) (bots.Edge, error) {
	pred, err := predicateFromDto(dto.Predicate)
	if err != nil {
		return bots.Edge{}, err
	}

	op, err := operationFromDto(dto.Operation)
	if err != nil {
		return bots.Edge{}, err
	}

	return bots.NewEdge(pred, bots.State(dto.To), op), nil
}

func batchEdgesFromDto(dto []Edge) ([]bots.Edge, error) {
	res := make([]bots.Edge, 0, len(dto))
	for _, edge := range dto {
		e, err := edgeFromDto(edge)
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func edgeToDto(e bots.Edge) Edge {
	return Edge{
		Predicate: predicateToDto(e.Predicate),
		To:        int(e.To()),
		Operation: operationToDto(e.Operation()),
	}
}

func batchEdgesToDto(edges []bots.Edge) []Edge {
	res := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		res = append(res, edgeToDto(edge))
	}
	return res
}

func messageFromDto(dto Message) (bots.Message, error) {
	return bots.NewMessage(dto.Text)
}

func messageToDto(m bots.Message) Message {
	return Message{
		Text: m.Text(),
	}
}

func batchMessagesFromDto(dto []Message) ([]bots.Message, error) {
	res := make([]bots.Message, 0, len(dto))
	for _, message := range dto {
		m, err := messageFromDto(message)
		if err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

func batchMessagesToDto(dto []bots.Message) []Message {
	res := make([]Message, len(dto))
	for i, message := range dto {
		res[i] = messageToDto(message)
	}
	return res
}

func batchOptionsFromDto(dto []string) []bots.Option {
	res := make([]bots.Option, len(dto))
	for i, option := range dto {
		res[i] = bots.Option(option)
	}
	return res
}

func batchOptionsToDto(dto []bots.Option) []string {
	res := make([]string, len(dto))
	for i, option := range dto {
		res[i] = string(option)
	}
	return res
}

func nodeFromDto(dto Node) (bots.Node, error) {
	edges, err := batchEdgesFromDto(dto.Edges)
	if err != nil {
		return bots.Node{}, err
	}

	messages, err := batchMessagesFromDto(dto.Messages)
	if err != nil {
		return bots.Node{}, err
	}

	options := batchOptionsFromDto(dto.Options)

	return bots.NewNode(bots.State(dto.State), dto.Title, edges, messages, options)
}

func batchNodesFromDto(dto []Node) ([]bots.Node, error) {
	res := make([]bots.Node, 0, len(dto))
	for _, node := range dto {
		n, err := nodeFromDto(node)
		if err != nil {
			return nil, err
		}
		res = append(res, n)
	}
	return res, nil
}

func nodeToDto(node bots.Node) Node {
	return Node{
		State:    int(node.State()),
		Title:    node.Title(),
		Edges:    batchEdgesToDto(node.Edges()),
		Messages: batchMessagesToDto(node.Messages()),
		Options:  batchOptionsToDto(node.Options()),
	}
}

func batchNodesToDto(nodes []bots.Node) []Node {
	res := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		res = append(res, nodeToDto(node))
	}
	return res
}

func entryFromDto(dto Entry) (bots.Entry, error) {
	return bots.NewEntry(bots.EntryKey(dto.Key), bots.State(dto.Start))
}

func batchEntriesFromDto(dto []Entry) ([]bots.Entry, error) {
	res := make([]bots.Entry, 0, len(dto))
	for _, entry := range dto {
		e, err := entryFromDto(entry)
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func entryToDto(entry bots.Entry) Entry {
	return Entry{
		Key:   string(entry.Key()),
		Start: int(entry.Start()),
	}
}

func batchEntriesToDto(entry []bots.Entry) []Entry {
	res := make([]Entry, 0, len(entry))
	for _, entry := range entry {
		res = append(res, entryToDto(entry))
	}
	return res
}

func scriptFromDto(dto Script) (bots.Script, error) {
	nodes, err := batchNodesFromDto(dto.Nodes)
	if err != nil {
		return bots.Script{}, err
	}

	entries, err := batchEntriesFromDto(dto.Entries)
	if err != nil {
		return bots.Script{}, err
	}

	return bots.NewScript(nodes, entries)
}

func scriptToDto(script bots.Script) Script {
	return Script{
		Nodes:   batchNodesToDto(script.Nodes()),
		Entries: batchEntriesToDto(script.Entries()),
	}
}

func botToDto(bot bots.Bot) Bot {
	script := scriptToDto(bot.Script())
	return Bot{
		ID:     string(bot.ID()),
		Token:  string(bot.Token()),
		Author: int64(bot.Author()),
		Script: script,
	}
}

func batchBotToDto(bs []bots.Bot) []Bot {
	res := make([]Bot, 0, len(bs))
	for _, bot := range bs {
		res = append(res, botToDto(bot))
	}
	return res
}

func threadToDto(thread *bots.Thread, username string) Thread {
	answers := make(map[int]Message)
	for state, msg := range thread.Answers() {
		answers[int(state)] = messageToDto(msg)
	}

	return Thread{
		ID:        string(thread.ID()),
		Key:       string(thread.Key()),
		StartedAt: thread.StartedAt(),
		Username:  username,
		Answers:   answers,
	}
}
