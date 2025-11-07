package bots

import (
	"slices"
)

// State есть состояние в контексте FSM и уникальный номер узла
// в пределах скрипта.
type State int

// Node есть минимальная структурная единица Script.
type Node struct {
	state State     // Собственный State узла.
	title string    // Заголовок узла
	edges []Edge    // Отсортированный по приоритету список исходящих рёбер.
	msgs  []Message // Список сообщений, который будет отправлен пользователю.
	opts  []Option  // Список кнопок-клавиатуры, которые будут отправлены с последним сообщением
}

// NewNode создаёт Node. msgs должно содержать как минимум одно BotMessage.
func NewNode(state State, title string, edges []Edge, msgs []Message, opts []Option) (Node, error) {
	if state < 0 {
		return Node{}, NewInvalidInputError(
			"invalid-node",
			"expected non-negative state",
		)
	}

	if title == "" {
		return Node{}, NewInvalidInputError(
			"invalid-node",
			"expected not empty title",
		)
	}

	if edges == nil {
		edges = make([]Edge, 0)
	}

	if len(msgs) == 0 {
		return Node{}, NewInvalidInputError(
			"invalid-node",
			"expected at least one message in node",
		)
	}

	if opts == nil {
		opts = make([]Option, 0)
	}

	return Node{
		state: state,
		title: title,
		edges: edges,
		msgs:  msgs,
		opts:  opts,
	}, nil
}

func MustNewNode(state State, title string, edges []Edge, msgs []Message, opts []Option) Node {
	n, err := NewNode(state, title, edges, msgs, opts)
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

// Messages возвращает сообщения в том виде, в котором они хранятся в узле.
func (n Node) Messages() []Message {
	return n.msgs
}

// BotMessages возвращает сообщения в том виде, в котором они будут отправлены пользователю.
// Если для узла заданы опции, последнее сообщение будет их содержать.
func (n Node) BotMessages() []BotMessage {
	res := make([]BotMessage, len(n.msgs))
	for i, msg := range n.msgs[:len(n.msgs)-1] {
		// Промежуточные сообщения не могут иметь опций ответа.
		res[i] = msg.PromoteToBotMessage(nil)
	}
	// Последнее сообщение гарантировано существует, т.к. len(n.msgs) > 0.
	// Добавляем к нему опции.
	res[len(res)-1] = n.msgs[len(n.msgs)-1].PromoteToBotMessage(n.opts)
	return res
}

func (n Node) Edges() []Edge {
	return n.edges
}

func (n Node) Options() []Option {
	return n.opts
}
