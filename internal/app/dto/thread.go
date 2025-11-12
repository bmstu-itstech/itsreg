package dto

import (
	"time"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Thread struct {
	ID        string
	Key       string
	StartedAt time.Time
	Username  string
	Answers   map[int]Message
}

func ThreadToDto(thread *bots.Thread, username string) Thread {
	answers := make(map[int]Message)
	for state, msg := range thread.Answers() {
		answers[state.Int()] = MessageToDTO(msg)
	}

	return Thread{
		ID:        string(thread.ID()),
		Key:       string(thread.Key()),
		StartedAt: thread.StartedAt(),
		Username:  username,
		Answers:   answers,
	}
}
