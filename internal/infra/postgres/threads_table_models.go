package postgres

import "time"

type threadTableRow struct {
	ID        string    `db:"id"`
	Key       string    `db:"key"`
	UserID    int64     `db:"user_id"`
	Username  string    `db:"username"`
	Timestamp time.Time `db:"ts"`
	Header    string    `db:"header"`
	Value     *string   `db:"value"`
}
