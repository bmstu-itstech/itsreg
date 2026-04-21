package bots

// Token есть Telegram токен для бота.
type Token string

func (t Token) IsZero() bool {
	return t == ""
}

func (t Token) String() string {
	return string(t)
}
