package bots

// Option есть доступная пользователю опция для выбора ответа.
// Для Telegram это ReplyKeyboardButton.
type Option string

type BotMessage struct {
	Message
	options []Option
}

func NewBotMessageWithoutOptions(text string) BotMessage {
	msg := NewMessage(text)
	return BotMessage{
		Message: msg,
		options: []Option{},
	}
}

func NewBotMessage(text string, opts []Option) (BotMessage, error) {
	msg := NewMessage(text)

	for _, opt := range opts {
		if opt == "" {
			return BotMessage{}, NewInvalidInputError(
				"invalid-bot-message-empty-option",
				"expected non-empty string options",
			)
		}
	}

	return BotMessage{
		Message: msg,
		options: opts[:],
	}, nil
}

func MustNewBotMessage(text string, opts []Option) BotMessage {
	m, err := NewBotMessage(text, opts)
	if err != nil {
		panic(err)
	}
	return m
}

func (m BotMessage) Options() []Option {
	return m.options[:]
}
