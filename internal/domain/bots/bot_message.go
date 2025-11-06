package bots

// Option есть доступная пользователю опция для выбора ответа.
// Для Telegram это ReplyKeyboardButton.
type Option string

type BotMessage struct {
	Message
	opts []Option
}

func (m BotMessage) Options() []Option {
	return m.opts[:]
}
