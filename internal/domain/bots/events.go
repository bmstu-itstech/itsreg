package bots

import "time"

type RunStartRequested struct {
	RunID RunID
	BotID BotID
	Time  time.Time
}

func (e RunStartRequested) EventName() string {
	return "run.start_requested"
}

func (e RunStartRequested) OccurredAt() time.Time {
	return e.Time
}

type RunRecoverRequested struct {
	RunID RunID
	Time  time.Time
}

func (e RunRecoverRequested) EventName() string {
	return "run.recover_requested"
}

func (e RunRecoverRequested) OccurredAt() time.Time {
	return e.Time
}

type RunStarted struct {
	RunID RunID
	BotID BotID
	Time  time.Time
}

func (e RunStarted) EventName() string {
	return "run.started"
}

func (e RunStarted) OccurredAt() time.Time {
	return e.Time
}

type RunFailed struct {
	RunID  RunID
	BotID  BotID
	ErrMsg string
	Time   time.Time
}

func (e RunFailed) EventName() string {
	return "run.failed"
}

func (e RunFailed) OccurredAt() time.Time {
	return e.Time
}

type RunStopped struct {
	RunID RunID
	BotID BotID
	Time  time.Time
}

func (e RunStopped) EventName() string {
	return "run.stopped"
}

func (e RunStopped) OccurredAt() time.Time {
	return e.Time
}
