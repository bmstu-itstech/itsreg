package bots

const messageMergeDelim = "\n"

type Message struct {
	text string
}

func NewMessage(text string) Message {
	return Message{
		text: text,
	}
}

// Text возвращает строго текст сообщения.
// Может быть пустым, например, если пользователь отправить файл.
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
