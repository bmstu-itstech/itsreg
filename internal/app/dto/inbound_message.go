package dto

type InboundMessage struct {
	BotID     string
	UserID    int64
	Username  *string
	Text      string
	IsCommand bool
}
