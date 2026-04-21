package dto

type Node struct {
	State    int
	Title    string
	Edges    []Edge
	Messages []Message
	Options  []string
}
