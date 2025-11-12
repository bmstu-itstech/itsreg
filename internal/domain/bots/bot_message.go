package bots

// Option есть доступная пользователю опция для выбора ответа.
// Для Telegram это ReplyKeyboardButton.
type Option struct {
	s string
}

func NewOption(s string) (Option, error) {
	if s == "" {
		return Option{}, NewInvalidInputError("option-empty-string", "expected not empty option string")
	}
	return Option{s: s}, nil
}

func MustNewOption(s string) Option {
	o, err := NewOption(s)
	if err != nil {
		panic(err)
	}
	return o
}

func (o Option) String() string {
	return o.s
}

type BotMessage struct {
	Message

	opts []Option
}

func (m BotMessage) Options() []Option {
	return m.opts
}
