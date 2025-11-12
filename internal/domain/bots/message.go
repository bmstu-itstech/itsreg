package bots

const messageMergeDelim = "\n"

type Message struct {
	text string
}

func NewMessage(text string) (Message, error) {
	if text == "" {
		return Message{}, NewInvalidInputError("message-empty-text", "expected not empty message text", "field", "text")
	}

	return Message{
		text: text,
	}, nil
}

func MustNewMessage(text string) Message {
	m, err := NewMessage(text)
	if err != nil {
		panic(err)
	}
	return m
}

// Text возвращает строго текст сообщения.
// В будущем, теоретически, может быть пустым, например, если пользователь отправил файл.
func (m Message) Text() string {
	return m.text
}

// String возвращает строковое представление сообщения.
// В отличие от Message.Text гарантируется, что строка не будет пустой.
func (m Message) String() string {
	return m.text
}

// Merge объединяет два сообщения.
// Как? Не должно иметь значения. Пока сообщения содержат только текст,
// объединять будем конкатенацией строк с разделителем messageMergeDelim.
func (m Message) Merge(o Message) Message {
	return Message{
		text: m.text + messageMergeDelim + o.text,
	}
}

// PromoteToBotMessage модифицирует сообщение для отправки его ботом.
func (m Message) PromoteToBotMessage(opts []Option) BotMessage {
	return BotMessage{
		Message: m,
		opts:    opts,
	}
}
