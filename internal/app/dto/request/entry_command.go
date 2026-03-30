package request

type EntryCommand struct {
	BotID    string
	UserID   int64
	Username string
	EntryKey string
}
