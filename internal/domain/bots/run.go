package bots

import (
	"errors"
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

var ErrIllegalStateTransition = errors.New("illegal state transition")

// Run представляет собой запуск бота на определённом токене. Содержит в себе всю
// информацию о запуске, включая статус, время запуска и остановки, а также
// возможную ошибку при выполнении.
//
// Жизненный цикл Run:
//  1. Создаётся в статусе StatusStarting с помощью NewRun.
//  2. Переходит в статус StatusActive при успешном запуске бота (метод Started).
//  3. Может перейти в статус StatusFailed при ошибке запуска, выполнения или
//     остановки (метод Failed).
//  4. Переходит в статус StatusStopping при запросе остановки (метод Stop).
//  5. Переходит в статус StatusStopped после успешной остановки (метод Stopped).
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

// Started переводит Run из статуса StatusStarting в StatusActive. Устанавливает время
// запуска и генерирует событие RunStarted. Если текущий статус не StatusStarting,
// возвращает ошибку ErrIllegalStateTransition.
func (r *Run) Started() error {
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

// Failed переводит Run в статус StatusFailed. Устанавливает время остановки,
// сохраняет сообщение об ошибке и генерирует событие RunFailed. Если текущий
// статус StatusStopped, то возвращает ошибку ErrIllegalStateTransition.
func (r *Run) Failed(msg string) error {
	if r.status == StatusStopped {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusStopped)
	}
	r.status = StatusFailed
	now := time.Now()
	r.stoppedAt = &now
	r.errorMsg = &msg
	r.events = append(r.events, RunFailed{
		RunID:  r.id,
		BotID:  r.botID,
		ErrMsg: msg,
		Time:   now,
	})
	return nil
}

// Stop переводит Run из статуса StatusActive в StatusStopping. Генерирует
// событие RunStopRequested. Если текущий статус не StatusActive, возвращает
// ошибку ErrIllegalStateTransition.
func (r *Run) Stop() error {
	if r.status != StatusActive {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusStopping)
	}
	r.status = StatusStopping
	now := time.Now()
	r.events = append(r.events, RunStopRequested{
		RunID: r.id,
		BotID: r.botID,
		Time:  now,
	})
	return nil
}

// Stopped переводит Run из статуса StatusStopping в StatusStopped. Устанавливает время
// остановки и генерирует событие RunStopped. Если текущий статус не StatusStopping,
// возвращает ошибку ErrIllegalStateTransition.
func (r *Run) Stopped() error {
	if r.status != StatusStopping {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, r.status, StatusStopped)
	}
	r.status = StatusStopped
	now := time.Now()
	r.stoppedAt = &now
	r.events = append(r.events, RunStopped{
		RunID: r.id,
		BotID: r.botID,
		Time:  now,
	})
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
