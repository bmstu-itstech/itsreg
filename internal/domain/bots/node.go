package bots

import (
	"slices"
)

// State есть состояние в контексте FSM и уникальный номер узла
// в пределах скрипта.
type State uint

// Node есть минимальная структурная единица Script.
type Node struct {
	state State        // Собственный State узла.
	title string       // Заголовок узла
	edges []Edge       // Отсортированный по приоритету список исходящих рёбер.
	msgs  []BotMessage // Список сообщений, который будет отправлен пользователю.
}

// NewNode создаёт Node. msgs должно содержать как минимум одно BotMessage.
func NewNode(state State, title string, edges []Edge, msgs []BotMessage) (Node, error) {
	if title == "" {
		return Node{}, NewInvalidInputError(
			"invalid-node",
			"expected not empty title",
		)
	}

	if len(msgs) == 0 {
		return Node{}, NewInvalidInputError(
			"invalid-node",
			"expected at least one message in node",
		)
	}

	return Node{
		state: state,
		title: title,
		edges: edges[:],
		msgs:  msgs[:],
	}, nil
}

func MustNewNode(state State, title string, edges []Edge, msgs []BotMessage) Node {
	n, err := NewNode(state, title, edges, msgs)
	if err != nil {
		panic(err)
	}
	return n
}

func (n Node) IsZero() bool {
	// Конструктор гарантирует, что msgs не будет nil.
	// Поэтому если msgs = nil, то сущность создана не через конструктор,
	// а значит, пустая.
	return n.msgs == nil
}

// Transition совершает условный переход по ребру с наивысшим приоритетом
// или возвращает false.
func (n Node) Transition(msg Message) (Edge, bool) {
	for _, edge := range n.edges {
		if edge.Match(msg) {
			return edge, true
		}
	}
	return Edge{}, false
}

// Children возвращает упорядоченное множество State дочерних узлов.
// Обычно используется для обхода графа.
func (n Node) Children() []State {
	children := make([]State, 0, len(n.edges))
	for _, edge := range n.edges {
		to := edge.To()
		// Повторные вхождения игнорируем
		if slices.Contains(children, to) {
			continue
		}
		children = append(children, to)
	}
	return children
}

func (n Node) State() State {
	return n.state
}

func (n Node) Title() string {
	return n.title
}

func (n Node) Messages() []BotMessage {
	return n.msgs[:]
}

func (n Node) Edges() []Edge {
	return n.edges[:]
}
