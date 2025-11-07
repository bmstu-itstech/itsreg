package bots

type Priority int

type Edge struct {
	Predicate

	to        State
	operation Operation
}

func NewEdge(pred Predicate, to State, op Operation) Edge {
	return Edge{
		Predicate: pred,
		to:        to,
		operation: op,
	}
}

func (e Edge) To() State {
	return e.to
}

func (e Edge) Operation() Operation {
	return e.operation
}
