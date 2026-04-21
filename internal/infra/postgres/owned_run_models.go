package postgres

import "time"

type ownedRunRow struct {
	ID        string     `db:"id"`
	OwnerID   int64      `db:"owner_id"`
	BotID     string     `db:"bot_id"`
	Status    string     `db:"status"`
	ErrorMsg  *string    `db:"error_msg"`
	StartedAt *time.Time `db:"started_at"`
	StoppedAt *time.Time `db:"stopped_at"`
}
