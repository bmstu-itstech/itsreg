package dto

import "time"

type Bot struct {
	ID        string
	OwnerID   int64
	ScriptID  string
	Desc      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
