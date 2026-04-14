package bots

const ErrorCodeOptionEmpty ErrorCode = "option-empty"

// Option есть доступная пользователю опция для выбора ответа.
// Для Telegram это ReplyKeyboardButton.
type Option struct {
	s string
}

func NewOption(s string) (Option, error) {
	if s == "" {
		return Option{}, NewValidationError(NewValidationErrorDetail(
			"value", ErrorCodeOptionEmpty, "option cannot be empty",
		))
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
