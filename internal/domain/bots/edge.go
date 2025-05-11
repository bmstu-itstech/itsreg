package bots

type Priority int

type Edge struct {
	Predicate
	to        State
	priority  Priority
	operation Operation
}

func NewEdge(pred Predicate, to State, pr Priority, op Operation) Edge {
	return Edge{
		Predicate: pred,
		to:        to,
		priority:  pr,
		operation: op,
	}
}

func (e Edge) To() State {
	return e.to
}

func (e Edge) Priority() Priority {
	return e.priority
}

func (e Edge) Operation() Operation {
	return e.operation
}

func CompareEdges(x, y Edge) int {
	return int(y.Priority() - x.Priority())
}
