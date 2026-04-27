package postgres

import "time"

type selectMailingsFilter struct {
	BotID  *string
	Status *string
}

type mailingRow struct {
	ID          string     `db:"id"`
	BotID       string     `db:"bot_id"`
	Name        string     `db:"name"`
	EntryKey    string     `db:"entry_key"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	StartedAt   *time.Time `db:"started_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

type mailingRecipientRow struct {
	MailingID string `db:"mailing_id"`
	UserID    int64  `db:"user_id"`
}

type mailingResultRow struct {
	MailingID string  `db:"mailing_id"`
	UserID    int64   `db:"user_id"`
	Success   bool    `db:"success"`
	ErrorMsg  *string `db:"error_msg"`
}
