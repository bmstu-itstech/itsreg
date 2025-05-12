package bots

// Option есть доступная пользователю опция для выбора ответа.
// Для Telegram это ReplyKeyboardButton.
type Option string

type BotMessage struct {
	Message
	options []Option
}

func NewBotMessage(text string, opts []Option) (BotMessage, error) {
	msg, err := NewMessage(text)
	if err != nil {
		return BotMessage{}, err
	}

	for _, opt := range opts {
		if opt == "" {
			return BotMessage{}, NewInvalidInputError(
				"invalid-bot-message",
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
