package bots

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNoStartedThread = errors.New("has no started thread")

type EntryNotFoundError struct {
	key EntryKey
}

func (e EntryNotFoundError) Error() string {
	return fmt.Sprintf("entry not found: %s", e.key)
}

// Script есть орграф с заданным множеством входных узлов.
type Script struct {
	nodes   map[State]Node
	entries map[EntryKey]Entry
}

func NewScript(_nodes []Node, _entries []Entry) (Script, error) {
	nodes := mapNodes(_nodes)
	entries := mapEntries(_entries)

	if err := checkConnectivity(nodes, entries); err != nil {
		return Script{}, err
	}

	return Script{
		nodes:   nodes,
		entries: entries,
	}, nil
}

func MustNewScript(_nodes []Node, _entries []Entry) Script {
	s, err := NewScript(_nodes, _entries)
	if err != nil {
		panic(err)
	}
	return s
}

func (s Script) IsZero() bool {
	// Достаточно быть пустому списку узлов, чтобы понять,
	// что скрипт был проинициализирован значениями по умолчанию
	return s.nodes == nil
}

func (s Script) Entry(prt *Participant, key EntryKey) ([]BotMessage, error) {
	entry, ok := s.entries[key]
	if !ok {
		return nil, EntryNotFoundError{key: key}
	}

	thread, err := prt.StartThread(entry)
	if err != nil {
		return nil, err
	}

	current, ok := s.nodes[thread.State()]
	if !ok {
		// Строго говоря, доменные правила запрещают появление такой ситуации, что
		// Participant будет иметь несуществующий state.
		return nil, fmt.Errorf("no bot node with state %d", thread.State())
	}

	return current.BotMessages(), nil
}

func (s Script) Process(prt *Participant, in Message) ([]BotMessage, error) {
	thread, ok := prt.CurrentThread()
	if !ok {
		return nil, ErrNoStartedThread
	}

	current, ok := s.nodes[thread.State()]
	if !ok {
		// Строго говоря, доменные правила запрещают появление такой ситуации, что
		// Participant будет иметь несуществующий state.
		return nil, fmt.Errorf("no bot node with state %d", thread.State())
	}

	edge, ok := current.Transition(in)
	if !ok {
		// Если сообщение не совпало ни с одним ребром, то ситуация не является
		// исключительной - ничего не происходит
		return nil, nil
	}
	edge.Operation().Apply(thread, in)

	nextState := edge.To()
	next, ok := s.nodes[nextState]
	if !ok {
		// Аналогично, схемой гарантируется, что следующий state будет существовать.
		return nil, fmt.Errorf("no bot node with state %d", nextState)
	}

	thread.StepTo(nextState)

	return next.BotMessages(), nil
}

func (s Script) Nodes() []Node {
	nodes := make([]Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (s Script) Entries() []Entry {
	entries := make([]Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	return entries
}

type color int

const (
	white color = iota // Узел не был пройден
	grey               // Узел в процессе обработки
	black              // Узел уже пройден
)

type coloredNode struct {
	Node

	Color color
}

func mapNodes(nodes []Node) map[State]Node {
	m := make(map[State]Node)
	for _, n := range nodes {
		m[n.State()] = n
	}
	return m
}

func mapEntries(entries []Entry) map[EntryKey]Entry {
	m := make(map[EntryKey]Entry)
	for _, e := range entries {
		m[e.Key()] = e
	}
	return m
}

func checkConnectivity(nodes map[State]Node, entries map[EntryKey]Entry) error {
	cns := coloredNodes(nodes)
	for _, entry := range entries {
		err := colorize(entry.Start(), cns)
		if err != nil {
			return err
		}
	}

	if ok, state := findWhiteNode(cns); ok {
		return NewInvalidInputError(
			"node-is-not-connected",
			fmt.Sprintf("node %d is connected, marked as an error", state),
			"state", strconv.Itoa(state.Int()),
		)
	}

	return nil
}

func coloredNodes(nodes map[State]Node) map[State]coloredNode {
	res := make(map[State]coloredNode)
	for state, node := range nodes {
		res[state] = coloredNode{node, white}
	}
	return res
}

func colorize(currentState State, nodes map[State]coloredNode) error {
	current, ok := nodes[currentState]
	if !ok {
		return NewInvalidInputError(
			"node-not-found",
			fmt.Sprintf("node %d is not found", currentState),
			"state", strconv.Itoa(currentState.Int()),
		)
	}

	dye(nodes, currentState, grey)

	for _, nextState := range current.Children() {
		next, o := nodes[nextState]
		if !o {
			return NewInvalidInputError(
				"node-not-found",
				fmt.Sprintf("node %d is not found", currentState),
				"state", strconv.Itoa(currentState.Int()),
			)
		}

		if next.Color == white {
			err := colorize(nextState, nodes)
			if err != nil {
				return err
			}
			dye(nodes, nextState, black)
		}
	}

	return nil
}

func dye(nodes map[State]coloredNode, state State, color color) {
	node := nodes[state]
	node.Color = color
	nodes[state] = node
}

func findWhiteNode(nodes map[State]coloredNode) (bool, State) {
	for state, node := range nodes {
		if node.Color == white {
			return true, state
		}
	}
	return false, ZeroState
}
