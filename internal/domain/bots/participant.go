package bots

import (
	"errors"
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
	id     ParticipantID
	thread *Thread // Активный тред для пользователя или nil
}

func NewParticipant(id ParticipantID) (*Participant, error) {
	if id.IsZero() {
		return nil, errors.New("id is zero")
	}

	return &Participant{
		id:     id,
		thread: nil,
	}, nil
}

func MustNewParticipant(id ParticipantID) *Participant {
	p, err := NewParticipant(id)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Participant) StartThread(entry Entry) (*Thread, error) {
	thread, err := NewThread(entry)
	if err != nil {
		return nil, err
	}
	p.thread = thread
	return thread, nil
}

func (p *Participant) ActiveThread() *Thread {
	return p.thread
}

func (p *Participant) ID() ParticipantID {
	return p.id
}

func UnmarshallParticipant(
	botID string,
	userID int64,
	thread *Thread,
) (*Participant, error) {
	if botID == "" {
		return nil, errors.New("BotID is empty")
	}

	if userID == 0 {
		return nil, errors.New("UserID is empty")
	}

	id := NewParticipantID(UserID(userID), BotID(botID))

	return &Participant{
		id:     id,
		thread: thread,
	}, nil
}
