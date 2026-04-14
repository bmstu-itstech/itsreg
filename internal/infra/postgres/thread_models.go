package postgres

import "time"

type threadRow struct {
	// PK(ID)
	ID        string    `db:"id"`
	BotID     string    `db:"bot_id"`
	UserID    int64     `db:"user_id"`
	Key       string    `db:"key"`
	State     int       `db:"state"`
	StartedAt time.Time `db:"started_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type answerRow struct {
	// PK(ThreadID, State)
	ThreadID string `db:"thread_id"`
	State    int    `db:"state"`
	Text     string `db:"text"`
}
