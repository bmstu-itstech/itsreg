package bots

type BotMessage struct {
	Message

	opts []Option
}

func (m BotMessage) Options() []Option {
	return m.opts
}
