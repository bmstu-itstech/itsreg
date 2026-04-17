package bots

import (
	"errors"
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

var ErrIllegalStateTransition = errors.New("illegal state transition")

type Run struct {
	id        RunID
	botID     BotID
	token     Token
	status    Status
	errorMsg  *string
	startedAt *time.Time
	stoppedAt *time.Time
	events    []event.Event
}

func NewRun(botID BotID, token Token) (*Run, error) {
	if botID == "" {
		return nil, errors.New("empty botID")
	}

	r := &Run{
		id:     NewRunID(),
		botID:  botID,
		token:  token,
		status: StatusStarting,
	}

	r.events = append(r.events, RunStartRequested{
		RunID: r.id,
		BotID: r.botID,
		Time:  time.Now(),
	})
	return r, nil
}

func RestoreRun(
	id RunID,
	botID BotID,
	token Token,
	status Status,
	errorMsg *string,
	startedAt *time.Time,
	stoppedAt *time.Time,
) (*Run, error) {
	if id.IsZero() {
		return nil, errors.New("zero run id")
	}

	if token.IsZero() {
		return nil, errors.New("zero token")
	}

	if botID.IsZero() {
		return nil, errors.New("zero botID")
	}

	if status.IsZero() {
		return nil, errors.New("zero status")
	}

	if startedAt != nil && startedAt.IsZero() {
		return nil, errors.New("zero startedAt")
	}

	if stoppedAt != nil && stoppedAt.IsZero() {
		return nil, errors.New("zero stoppedAt")
	}

	return &Run{
		id:        id,
		botID:     botID,
		token:     token,
		status:    status,
		errorMsg:  errorMsg,
		startedAt: startedAt,
		stoppedAt: stoppedAt,
	}, nil
}

func (r *Run) PullEvents() []event.Event {
	ev := r.events
	r.events = nil
	return ev
}

func (r *Run) Start() error {
	if r.status != StatusStarting {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusActive)
	}
	r.status = StatusActive
	now := time.Now()
	r.startedAt = &now
	r.events = append(r.events, RunStarted{
		RunID: r.id,
		BotID: r.botID,
		Time:  time.Now(),
	})
	return nil
}

func (r *Run) Fail(msg string) error {
	if r.status != StatusStarting && r.status != StatusActive {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusFailed)
	}
	r.status = StatusFailed
	t := time.Now()
	r.stoppedAt = &t
	r.errorMsg = &msg
	r.events = append(r.events, RunFailed{
		RunID:  r.id,
		BotID:  r.botID,
		ErrMsg: msg,
		Time:   t,
	})
	return nil
}

func (r *Run) Stop() error {
	if r.status != StatusActive {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusStopped)
	}
	r.status = StatusStopped
	t := time.Now()
	r.stoppedAt = &t
	return nil
}

func (r *Run) ID() RunID {
	return r.id
}

func (r *Run) BotID() BotID {
	return r.botID
}

func (r *Run) Token() Token {
	return r.token
}

func (r *Run) Status() Status {
	return r.status
}

func (r *Run) ErrorMsg() *string {
	return r.errorMsg
}

func (r *Run) StartedAt() *time.Time {
	return r.startedAt
}

func (r *Run) StoppedAt() *time.Time {
	return r.stoppedAt
}
