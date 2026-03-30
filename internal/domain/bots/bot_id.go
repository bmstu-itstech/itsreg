package bots

// BotID есть уникальный идентификатор бота.
type BotID string

func (b BotID) IsZero() bool {
	return b == ""
}

func (b BotID) String() string {
	return string(b)
}
