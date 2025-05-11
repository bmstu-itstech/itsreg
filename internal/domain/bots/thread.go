package bots

import (
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg-bots/pkg/uuid"
)

type ThreadId string

// Thread есть цепочка ответов Participant от Entry до конечного State или
// до следующего Entry.
type Thread struct {
	id        ThreadId
	key       EntryKey
	state     State
	answers   map[State]Message
	startedAt time.Time
}

func NewThread(entry Entry) (*Thread, error) {
	if entry.IsZero() {
		return nil, fmt.Errorf("entry is empty")
	}

	return &Thread{
		id:        ThreadId(uuid.Generate()),
		key:       entry.Key(),
		state:     entry.Start(),
		answers:   make(map[State]Message),
		startedAt: time.Now(),
	}, nil
}

func MustNewThread(entry Entry) *Thread {
	t, err := NewThread(entry)
	if err != nil {
		panic(err)
	}
	return t
}

func (t *Thread) StepTo(to State) {
	t.state = to
}

// SaveAnswer сохраняет Message пользователя для текущего состояния.
// Если уже существует ответ для данного состояния, перезаписывает его.
func (t *Thread) SaveAnswer(ans Message) {
	t.answers[t.state] = ans
}

// AppendAnswer сохраняет Message пользователя для текущего состояния.
// Если уже существует ответ для данного состояния, то объединяет новое
// сообщение с предыдущим через метод Message.Merge.
func (t *Thread) AppendAnswer(ans Message) {
	if saved, ok := t.answers[t.state]; ok {
		t.answers[t.state] = saved.Merge(ans)
	} else {
		t.answers[t.state] = ans
	}
}

func (t *Thread) Id() ThreadId {
	return t.id
}

func (t *Thread) Key() EntryKey {
	return t.key
}

func (t *Thread) State() State {
	return t.state
}

func (t *Thread) Answers() map[State]Message {
	return t.answers
}

func (t *Thread) StartedAt() time.Time {
	return t.startedAt
}
