package bots

import (
	"errors"
	"time"
)

// Thread есть цепочка ответов пользователя от Entry до конечного State или
// до следующего Entry.
type Thread struct {
	id        ThreadID
	botID     BotID
	userID    UserID
	key       EntryKey
	state     State
	answers   map[State]Message
	startedAt time.Time
}

func NewThread(botID BotID, userID UserID, entry Entry) (*Thread, error) {
	if botID.IsZero() {
		return nil, errors.New("botID is zero")
	}

	if userID.IsZero() {
		return nil, errors.New("userID is zero")
	}

	if entry.IsZero() {
		return nil, errors.New("entry is empty")
	}

	return &Thread{
		id:        NewThreadID(),
		botID:     botID,
		userID:    userID,
		key:       entry.Key(),
		state:     entry.Start(),
		answers:   make(map[State]Message),
		startedAt: time.Now(),
	}, nil
}

func MustNewThread(botID BotID, userID UserID, entry Entry) *Thread {
	t, err := NewThread(botID, userID, entry)
	if err != nil {
		panic(err)
	}
	return t
}

func RestoreThread(
	id ThreadID,
	botID BotID,
	userID UserID,
	key EntryKey,
	state State,
	answers map[State]Message,
	startedAt time.Time,
) (*Thread, error) {
	if id.IsZero() {
		return nil, errors.New("id is empty")
	}

	if key == "" {
		return nil, errors.New("key is empty")
	}

	if answers == nil {
		answers = make(map[State]Message)
	}

	if startedAt.IsZero() {
		return nil, errors.New("startedAt is empty")
	}

	return &Thread{
		id:        id,
		botID:     botID,
		userID:    userID,
		key:       key,
		state:     state,
		answers:   answers,
		startedAt: startedAt,
	}, nil
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

func (t *Thread) BotID() BotID {
	return t.botID
}

func (t *Thread) UserID() UserID {
	return t.userID
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
