package bots

import (
	"errors"
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

const (
	ErrorCodeScriptNodeIsNotConnected shared.ErrorCode = "script-node-is-not-connected"
	ErrorCodeScriptNodeNotFound       shared.ErrorCode = "script-node-not-found"
)

var (
	ErrEntryNotFound = errors.New("entry not found")
	ErrScriptDeleted = errors.New("bot deleted")
)

// Script есть агрегат, содержащий информацию о текущем сценарии некоторого бота, и представляет
// собой конечный автомат (орграф).
// Требуется, чтобы все узлы были достижимы хотя бы из одного Entry.
type Script struct {
	id        ScriptID
	ownerID   UserID
	desc      string
	nodes     map[State]Node
	entries   map[EntryKey]Entry
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewScript(
	ownerID UserID,
	desc string,
	_nodes []Node,
	_entries []Entry,
) (*Script, error) {
	if ownerID.IsZero() {
		return nil, errors.New("zero owner id")
	}

	nodes := mapNodes(_nodes)
	entries := mapEntries(_entries)

	if err := checkConnectivity(nodes, entries); err != nil {
		return nil, err
	}

	return &Script{
		id:        NewScriptID(),
		ownerID:   ownerID,
		desc:      desc,
		nodes:     nodes,
		entries:   entries,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		deletedAt: nil,
	}, nil
}

func MustNewScript(ownerID UserID, desc string, _nodes []Node, _entries []Entry) *Script {
	s, err := NewScript(ownerID, desc, _nodes, _entries)
	if err != nil {
		panic(err)
	}
	return s
}

func RestoreScript(
	id ScriptID,
	ownerID UserID,
	desc string,
	_nodes []Node,
	_entries []Entry,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Script, error) {
	if id.IsZero() {
		return nil, errors.New("zero script id")
	}

	if ownerID.IsZero() {
		return nil, errors.New("zero owner id")
	}

	if createdAt.IsZero() {
		return nil, errors.New("zero script created")
	}

	if updatedAt.IsZero() {
		return nil, errors.New("zero script updated")
	}

	if deletedAt != nil && deletedAt.IsZero() {
		return nil, errors.New("zero script deleted")
	}

	nodes := mapNodes(_nodes)
	entries := mapEntries(_entries)

	return &Script{
		id:        id,
		ownerID:   ownerID,
		desc:      desc,
		nodes:     nodes,
		entries:   entries,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}, nil
}

func (s *Script) Delete() error {
	if s.Deleted() {
		return fmt.Errorf("cannot delete script: %w", ErrScriptDeleted)
	}
	s.updatedAt = time.Now()
	t := time.Now()
	s.deletedAt = &t
	return nil
}

func (s *Script) Replace(desc string, _nodes []Node, _entries []Entry) error {
	if s.Deleted() {
		return fmt.Errorf("cannot replace script: %w", ErrScriptDeleted)
	}

	nodes := mapNodes(_nodes)
	entries := mapEntries(_entries)

	if err := checkConnectivity(nodes, entries); err != nil {
		return err
	}

	s.desc = desc
	s.nodes = nodes
	s.entries = entries
	s.updatedAt = time.Now()

	return nil
}

func (s *Script) Entry(botID BotID, userID UserID, key EntryKey) (*Thread, []BotMessage, error) {
	if s.Deleted() {
		return nil, nil, fmt.Errorf("cannot entry script: %w", ErrScriptDeleted)
	}

	entry, ok := s.entries[key]
	if !ok {
		return nil, nil, fmt.Errorf("%w: %s", ErrEntryNotFound, key)
	}

	thread, err := NewThread(botID, userID, entry)
	if err != nil {
		return nil, nil, err
	}

	current, ok := s.nodes[thread.State()]
	if !ok {
		// Строго говоря, доменные правила запрещают появление такой ситуации, что
		// Thread будет иметь несуществующий state.
		return nil, nil, fmt.Errorf("no bot node with state %d", thread.State())
	}

	return thread, current.BotMessages(), nil
}

func (s *Script) Process(thread *Thread, in Message) ([]BotMessage, error) {
	if s.Deleted() {
		return nil, fmt.Errorf("cannot process script: %w", ErrScriptDeleted)
	}

	current, ok := s.nodes[thread.State()]
	if !ok {
		// Строго говоря, доменные правила запрещают появление такой ситуации, что
		// Thread будет иметь несуществующий state.
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

func (s *Script) EnsureActive() error {
	if s.Deleted() {
		return ErrScriptDeleted
	}
	return nil
}

func (s *Script) EnsureOwnedBy(userID UserID) error {
	if s.ownerID != userID {
		return shared.ErrPermissionDenied
	}
	return nil
}

func (s *Script) ID() ScriptID {
	return s.id
}

func (s *Script) OwnerID() UserID {
	return s.ownerID
}

func (s *Script) Desc() string {
	return s.desc
}

func (s *Script) Nodes() []Node {
	nodes := make([]Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (s *Script) Entries() []Entry {
	entries := make([]Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	return entries
}

func (s *Script) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Script) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Script) DeletedAt() *time.Time {
	return s.deletedAt
}

func (s *Script) Deleted() bool {
	return s.deletedAt != nil
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
	for key, entry := range entries {
		_, ok := nodes[entry.Start()]
		if !ok {
			return shared.NewValidationError(shared.NewValidationErrorDetail(
				"nodes",
				ErrorCodeScriptNodeNotFound,
				fmt.Sprintf("an entry with key=%q references to non-existent node[%d]", key, entry.Start().Int()),
			))
		}
		err := colorize(entry.Start(), cns)
		if err != nil {
			return err
		}
	}

	whiteNodes := filterWhiteNodes(cns)
	if len(whiteNodes) > 0 {
		details := make([]shared.ValidationErrorDetail, 0, len(whiteNodes))
		for _, node := range whiteNodes {
			details = append(details, shared.NewValidationErrorDetail(
				"nodes", ErrorCodeScriptNodeIsNotConnected,
				fmt.Sprintf("nodes[%d] is not connected", node.State().Int()),
			))
		}
		return shared.NewValidationError(details...)
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

// Функция colorize раскрашивает nodes по следующим правилам:
// 1. Закрашивает currentState в серый цвет.
// 2. Если смежный узел - белый, то рекурсивно вызывает colorize для него, а затем закрашивает его в черный цвет.
// Узел с currentState должен быть представлен в nodes.
func colorize(currentState State, nodes map[State]coloredNode) error {
	current, ok := nodes[currentState]
	if !ok {
		// Существование currentState должно проверяться до вызова функции
		return fmt.Errorf("nodes[%d] not found", currentState)
	}

	dye(nodes, currentState, grey)

	for _, nextState := range current.Children() {
		next, o := nodes[nextState]
		if !o {
			return shared.NewValidationError(shared.NewValidationErrorDetail(
				"nodes",
				ErrorCodeScriptNodeNotFound,
				fmt.Sprintf("nodes[%d] references to non-existent nodes[%d]", current.State().Int(), nextState.Int()),
			))
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

func filterWhiteNodes(nodes map[State]coloredNode) map[State]coloredNode {
	res := make(map[State]coloredNode)
	for state, node := range nodes {
		if node.Color == white {
			res[state] = node
		}
	}
	return res
}
