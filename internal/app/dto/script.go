package dto

import "time"

type Script struct {
	ID        string
	OwnerID   int64
	Desc      string
	Nodes     []Node
	Entries   []Entry
	CreatedAt time.Time
	UpdatedAt time.Time
}
