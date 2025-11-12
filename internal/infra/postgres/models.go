package postgres

import "time"

type botRow struct {
	// PK (ID)
	ID        string    `db:"id"`
	Token     string    `db:"token"`
	Author    int64     `db:"author"`
	CreatedAt time.Time `db:"created_at"`
}

type entryRow struct {
	// PK (BotID, Key)
	BotID string `db:"bot_id"`
	Key   string `db:"key"`
	Start int    `db:"start"`
}

func entryIdentity(lhs, rhs entryRow) bool {
	return lhs.BotID == rhs.BotID && lhs.Key == rhs.Key
}

type nodeRow struct {
	// PK (BotID, State)
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Title string `db:"title"`
}

func nodeIdentity(lhs, rhs nodeRow) bool {
	return lhs.BotID == rhs.BotID && lhs.State == rhs.State
}

type edgeRow struct {
	BotID     string `db:"bot_id"`
	State     int    `db:"state"`
	ToState   int    `db:"to_state"`
	Operation string `db:"operation"`
	PredType  string `db:"pred_type"`
	PredData  string `db:"pred_data"`
}

type messageRow struct {
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Text  string `db:"text"`
}

type optionRow struct {
	BotID string `db:"bot_id"`
	State int    `db:"state"`
	Text  string `db:"text"`
}

type participantRow struct {
	// PK(BotID, UserID)
	BotID        string  `db:"bot_id"`
	UserID       int64   `db:"user_id"`
	ActiveThread *string `db:"active_thread"`
}

type threadRow struct {
	// PK(ID)
	ID        string    `db:"id"`
	BotID     string    `db:"bot_id"`
	UserID    int64     `db:"user_id"`
	Key       string    `db:"key"`
	State     int       `db:"state"`
	StartedAt time.Time `db:"started_at"`
}

type answerRow struct {
	// PK(ThreadID)
	ThreadID string `db:"thread_id"`
	State    int    `db:"state"`
	Text     string `db:"text"`
}

func answerIdentity(lhs, rhs answerRow) bool {
	return lhs.ThreadID == rhs.ThreadID && lhs.State == rhs.State
}
