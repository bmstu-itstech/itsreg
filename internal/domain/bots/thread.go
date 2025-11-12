package bots

import (
	"errors"
	"maps"
	"time"

	"github.com/bmstu-itstech/itsreg-bots/pkg/uuid"
)

type ThreadID string

// Thread есть цепочка ответов Participant от Entry до конечного State или
// до следующего Entry.
type Thread struct {
	id        ThreadID
	key       EntryKey
	state     State
	answers   map[State]Message
	startedAt time.Time
}

func NewThread(entry Entry) (*Thread, error) {
	if entry.IsZero() {
		return nil, errors.New("entry is empty")
	}

	return &Thread{
		id:        ThreadID(uuid.Generate()),
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

func (t *Thread) Clone() *Thread {
	return &Thread{
		id:        t.id,
		key:       t.key,
		state:     t.state,
		answers:   maps.Clone(t.answers),
		startedAt: t.startedAt,
	}
}

func (t *Thread) Equals(other *Thread) bool {
	return t.id == other.id &&
		t.key == other.key &&
		t.state == other.state &&
		maps.Equal(t.answers, other.answers) &&
		t.startedAt.Equal(other.startedAt)
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

func (t *Thread) ID() ThreadID {
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

type BotThread struct {
	thread *Thread
	botID  BotID
	userID UserID
}

func NewBotThread(thread *Thread, botID BotID, userID UserID) BotThread {
	return BotThread{
		thread: thread,
		botID:  botID,
		userID: userID,
	}
}

func (bt *BotThread) BotID() BotID {
	return bt.botID
}

func (bt *BotThread) UserID() UserID {
	return bt.userID
}

func (bt *BotThread) Thread() *Thread {
	return bt.thread
}

func UnmarshallThread(
	id string,
	key string,
	state int,
	answers map[State]Message,
	startedAt time.Time,
) (*Thread, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if key == "" {
		return nil, errors.New("key is empty")
	}

	s, err := NewState(state)
	if err != nil {
		return nil, err
	}

	if answers == nil {
		answers = make(map[State]Message)
	}

	if startedAt.IsZero() {
		return nil, errors.New("startedAt is empty")
	}

	return &Thread{
		id:        ThreadID(id),
		key:       EntryKey(key),
		state:     s,
		answers:   answers,
		startedAt: startedAt,
	}, nil
}
