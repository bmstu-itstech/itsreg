package dto

import "time"

type OwnedRun struct {
	ID        string
	OwnerID   int64
	BotID     string
	Status    string
	ErrorMsg  *string
	StartedAt *time.Time
	StoppedAt *time.Time
}
