package dto

import "time"

type Run struct {
	ID        string
	BotID     string
	Status    string
	ErrorMsg  *string
	StartedAt *time.Time
	StoppedAt *time.Time
}
