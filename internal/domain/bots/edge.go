package bots

import (
	"fmt"
	"regexp"
)

type Priority int

type Edge interface {
	// Match возвращает true если сообщение удовлетворяет предикату ребра.
	Match(s Message) bool

	// Priority задаёт порядок обхода рёбер.
	Priority() Priority

	// To есть следующий State узла.
	To() State

	// Operation есть действие, выполняемое над пользователем после прохода по ребру.
	Operation() Operation
}

func CompareEdges(x, y Edge) int {
	return int(y.Priority() - x.Priority())
}

type regexpEdge struct {
	cond      regexp.Regexp
	to        State
	priority  Priority
	operation Operation
}

func NewRegexpEdge(cond string, to State, pr Priority, op Operation) (Edge, error) {
	exp, err := regexp.Compile(cond)
	if err != nil {
		return nil, NewInvalidInputError(
			"invalid-regexp-edge",
			fmt.Sprintf("failed to compile regexp edge: %s", err.Error()),
		)
	}

	return regexpEdge{
		cond:      *exp,
		to:        to,
		priority:  pr,
		operation: op,
	}, nil
}

func MustNewRegexpEdge(cond string, to State, pr Priority, op Operation) Edge {
	e, err := NewRegexpEdge(cond, to, pr, op)
	if err != nil {
		panic(err)
	}
	return e
}

func (e regexpEdge) Match(s Message) bool {
	return e.cond.MatchString(s.Text())
}

func (e regexpEdge) To() State {
	return e.to
}

func (e regexpEdge) Priority() Priority {
	return e.priority
}

func (e regexpEdge) Operation() Operation {
	return e.operation
}
