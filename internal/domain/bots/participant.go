package bots

import (
	"errors"
	"fmt"
	"maps"
)

type UserID int64

type ParticipantID struct {
	userID UserID
	botID  BotID
}

func NewParticipantID(userID UserID, botID BotID) ParticipantID {
	return ParticipantID{
		userID: userID,
		botID:  botID,
	}
}

func (id ParticipantID) IsZero() bool {
	return id.botID == ""
}

func (id ParticipantID) UserID() UserID {
	return id.userID
}

func (id ParticipantID) BotID() BotID {
	return id.botID
}

type Participant struct {
	id              ParticipantID
	threads         map[ThreadID]*Thread
	currentThreadID *ThreadID
}

func NewParticipant(id ParticipantID) (*Participant, error) {
	if id.IsZero() {
		return nil, errors.New("id is zero")
	}

	return &Participant{
		id:              id,
		threads:         make(map[ThreadID]*Thread),
		currentThreadID: nil,
	}, nil
}

func MustNewParticipant(id ParticipantID) *Participant {
	p, err := NewParticipant(id)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Participant) Clone() *Participant {
	threads := make(map[ThreadID]*Thread)
	for k, v := range p.threads {
		threads[k] = v.Clone()
	}
	return &Participant{
		id:              p.id,
		threads:         threads,
		currentThreadID: p.currentThreadID, // Значение по указателю не изменяется - аналог Optional; поэтому можно
		// не копировать значение по адресу
	}
}

func (p *Participant) Equals(other *Participant) bool {
	return p.id == other.id &&
		equalThreadMaps(p.threads, other.threads) &&
		equalThreadIDs(p.currentThreadID, other.currentThreadID)
}

func equalThreadIDs(a, b *ThreadID) bool {
	if a != nil && b != nil {
		return *a == *b
	}
	return a == b
}

func equalThreadMaps(a, b map[ThreadID]*Thread) bool {
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
	id := thread.ID()
	p.threads[id] = thread
	p.currentThreadID = &id
	return thread, nil
}

func (p *Participant) CurrentThread() (*Thread, bool) {
	if p.currentThreadID == nil {
		return nil, false
	}
	thread := p.threads[*p.currentThreadID] // Гарантировано есть
	return thread, true
}

func (p *Participant) ID() ParticipantID {
	return p.id
}

func UnmarshallParticipant(
	botID string,
	userID int64,
	_threads []*Thread,
	_currentThreadID *string,
) (*Participant, error) {
	if botID == "" {
		return nil, errors.New("BotID is empty")
	}

	if userID == 0 {
		return nil, errors.New("UserID is empty")
	}

	threads := make(map[ThreadID]*Thread, len(_threads))
	for _, thread := range _threads {
		threads[thread.ID()] = thread
	}

	if _currentThreadID != nil && *_currentThreadID == "" {
		return nil, errors.New("currentThreadID is not null and empty")
	}

	id := NewParticipantID(UserID(userID), BotID(botID))

	var currentThreadID *ThreadID
	if _currentThreadID != nil {
		t := ThreadID(*_currentThreadID)
		if _, ok := threads[t]; !ok {
			return nil, fmt.Errorf("unknown thread: %s", t)
		}
		currentThreadID = &t
	}

	return &Participant{
		id:              id,
		threads:         threads,
		currentThreadID: currentThreadID,
	}, nil
}
