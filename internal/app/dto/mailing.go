package dto

import "time"

type Mailing struct {
	ID           string
	BotID        string
	Name         string
	EntryKey     string
	Status       string
	SuccessCount int
	FailCount    int
	PendingCount int
	TotalCount   int
	Recipients   []int64
	Results      []UserMailingResult
	CreatedAt    time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
}
