package request

type MailingCommand struct {
	BotID    string
	EntryKey string
	Users    []int64
}
