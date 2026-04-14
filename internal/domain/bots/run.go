package bots

import (
	"errors"
	"fmt"
	"time"
)

var ErrIllegalStateTransition = errors.New("illegal state transition")

type Run struct {
	id        RunID
	botID     BotID
	status    Status
	errorMsg  *string
	startedAt *time.Time
	stoppedAt *time.Time
}

func NewRun(botID BotID) (*Run, error) {
	if botID == "" {
		return nil, errors.New("empty botID")
	}

	return &Run{
		id:        NewRunID(),
		botID:     botID,
		status:    Starting,
		errorMsg:  nil,
		startedAt: nil,
		stoppedAt: nil,
	}, nil
}

func RestoreRun(
	id RunID,
	botID BotID,
	status Status,
	errorMsg *string,
	startedAt *time.Time,
	stoppedAt *time.Time,
) (*Run, error) {
	if id.IsZero() {
		return nil, errors.New("zero run id")
	}

	if botID.IsZero() {
		return nil, errors.New("zero botID")
	}

	if status.IsZero() {
		return nil, errors.New("zero status")
	}

	return &Run{
		id:        id,
		botID:     botID,
		status:    status,
		errorMsg:  errorMsg,
		startedAt: startedAt,
		stoppedAt: stoppedAt,
	}, nil
}

func (r *Run) Start() error {
	if r.status != Starting {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, Active)
	}
	r.status = Active
	t := time.Now()
	r.startedAt = &t
	return nil
}

func (r *Run) Fail(msg string) error {
	if r.status != Starting && r.status != Active {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, Failed)
	}
	r.status = Failed
	t := time.Now()
	r.stoppedAt = &t
	r.errorMsg = &msg
	return nil
}

func (r *Run) Stop() error {
	if r.status != Active {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, Stopped)
	}
	r.status = Stopped
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
