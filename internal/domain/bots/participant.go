package bots

import (
	"errors"
	"fmt"
	"maps"
)

type UserId int64

type ParticipantId struct {
	userId UserId
	botId  BotId
}

func NewParticipantId(userId UserId, botId BotId) ParticipantId {
	return ParticipantId{
		userId: userId,
		botId:  botId,
	}
}

func (id ParticipantId) IsZero() bool {
	return id.botId == ""
}

func (id ParticipantId) UserId() UserId {
	return id.userId
}

func (id ParticipantId) BotId() BotId {
	return id.botId
}

type Participant struct {
	id              ParticipantId
	threads         map[ThreadId]*Thread
	currentThreadId *ThreadId
}

func NewParticipant(id ParticipantId) (*Participant, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("id is zero")
	}

	return &Participant{
		id:              id,
		threads:         make(map[ThreadId]*Thread),
		currentThreadId: nil,
	}, nil
}

func MustNewParticipant(id ParticipantId) *Participant {
	p, err := NewParticipant(id)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Participant) Clone() *Participant {
	threads := make(map[ThreadId]*Thread)
	for k, v := range p.threads {
		threads[k] = v.Clone()
	}
	return &Participant{
		id:              p.id,
		threads:         threads,
		currentThreadId: p.currentThreadId, // Значение по указателю не изменяется - аналог Optional; поэтому можно
		// не копировать значение по адресу
	}
}

func (p *Participant) Equals(other *Participant) bool {
	return p.id == other.id &&
		equalThreadMaps(p.threads, other.threads) &&
		equalThreadIds(p.currentThreadId, other.currentThreadId)
}

func equalThreadIds(a, b *ThreadId) bool {
	if a != nil && b != nil {
		return *a == *b
	}
	return a == b
}

func equalThreadMaps(a, b map[ThreadId]*Thread) bool {
	return maps.EqualFunc(a, b, func(t1 *Thread, t2 *Thread) bool {
		return t1.Equals(t2)
	})
}

func (p *Participant) Threads() []*Thread {
	threads := make([]*Thread, 0, len(p.threads))
	for _, thread := range p.threads {
		threads = append(threads, thread.Clone())
	}
	return threads
}

func (p *Participant) StartThread(entry Entry) (*Thread, error) {
	thread, err := NewThread(entry)
	if err != nil {
		return nil, err
	}
	id := thread.Id()
	p.threads[id] = thread
	p.currentThreadId = &id
	return thread, nil
}

func (p *Participant) CurrentThread() (*Thread, bool) {
	if p.currentThreadId == nil {
		return nil, false
	}
	thread := p.threads[*p.currentThreadId] // Гарантировано есть
	return thread, true
}

func (p *Participant) Id() ParticipantId {
	return p.id
}

func UnmarshallParticipant(
	botId string,
	userId int64,
	_threads []*Thread,
	_currentThreadId *string,
) (*Participant, error) {
	if botId == "" {
		return nil, errors.New("botId is empty")
	}

	if userId == 0 {
		return nil, errors.New("userId is empty")
	}

	threads := make(map[ThreadId]*Thread, len(_threads))
	for _, thread := range _threads {
		threads[thread.Id()] = thread
	}

	if _currentThreadId != nil && *_currentThreadId == "" {
		return nil, errors.New("currentThreadId is not null and empty")
	}

	id := NewParticipantId(UserId(userId), BotId(botId))

	var currentThreadId *ThreadId
	if _currentThreadId != nil {
		t := ThreadId(*_currentThreadId)
		if _, ok := threads[t]; !ok {
			return nil, fmt.Errorf("unknown thread: %s", t)
		}
		currentThreadId = &t
	}

	return &Participant{
		id:              id,
		threads:         threads,
		currentThreadId: currentThreadId,
	}, nil
}
