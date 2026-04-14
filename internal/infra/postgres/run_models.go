package postgres

import "time"

type runRow struct {
	// PK (ID)
	ID        string     `db:"id"`
	BotID     string     `db:"bot_id"`
	Status    string     `db:"status"`
	ErrorMsg  *string    `db:"error_msg"`
	StartedAt *time.Time `db:"started_at"`
	StoppedAt *time.Time `db:"stopped_at"`
}
