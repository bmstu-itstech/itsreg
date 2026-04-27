package postgres

import "time"

type ownedMailingModel struct {
	ID           string     `db:"id"`
	OwnerID      int64      `db:"owner_id"`
	BotID        string     `db:"bot_id"`
	Name         string     `db:"name"`
	EntryKey     string     `db:"entry_key"`
	Status       string     `db:"status"`
	SuccessCount int        `db:"success_count"`
	FailCount    int        `db:"fail_count"`
	PendingCount int        `db:"pending_count"`
	TotalCount   int        `db:"total_count"`
	CreatedAt    time.Time  `db:"created_at"`
	StartedAt    *time.Time `db:"started_at"`
	CompletedAt  *time.Time `db:"completed_at"`
}
