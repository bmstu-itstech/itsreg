package bots

import "fmt"

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

func (p *Participant) Threads() []Thread {
	threads := make([]Thread, 0, len(p.threads))
	for _, thread := range p.threads {
		threads = append(threads, *thread)
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
