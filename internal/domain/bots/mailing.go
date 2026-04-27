package bots

import (
	"errors"
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
)

const (
	ErrorCodeMailingEmptyName       shared.ErrorCode = "mailing-empty-name"
	ErrorCodeMailingEmptyRecipients shared.ErrorCode = "mailing-empty-recipients"
)

var (
	ErrUserNotInRecipients = errors.New("user is not in recipients")
	ErrMailingIsNotStarted = errors.New("mailing is not started")
)

// Mailing представляет собой рассылку сообщений пользователям от имени бота.
// Содержит в себе всю информацию о рассылке, включая статус, список
// получателей, результаты отправки и время создания, начала и завершения
// рассылки.
//
// Жизненный цикл Mailing:
//  1. Создаётся в статусе MailingStatusScheduled с помощью NewMailing.
//  2. Переходит в статус MailingStatusStarted при начале рассылки (метод MarkStarted).
//  3. Может перейти в статус MailingStatusFailed при ошибке выполнения (метод MarkFailed).
//  4. Переходит в статус MailingStatusCompleted после успешной отправки сообщений всем пользователям.
type Mailing struct {
	id          MailingID
	botID       BotID
	name        string
	entryKey    EntryKey
	status      MailingStatus
	recipients  map[UserID]bool
	results     map[UserID]UserMailingResult
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	events      []event.Event
}

func NewMailing(botID BotID, name string, entryKey EntryKey, recipients []UserID) (*Mailing, error) {
	if botID.IsZero() {
		return nil, errors.New("bot ID is zero")
	}

	if name == "" {
		return nil, shared.NewValidationError(shared.NewValidationErrorDetail(
			"name", ErrorCodeMailingEmptyName, "mailing name is required",
		))
	}

	if entryKey.IsZero() {
		return nil, errors.New("entry key is zero")
	}

	if len(recipients) == 0 {
		return nil, shared.NewValidationError(shared.NewValidationErrorDetail(
			"recipients", ErrorCodeMailingEmptyRecipients, "mailing should contain at least one recipient",
		))
	}

	recsMapped := mapUserIDs(recipients)

	m := &Mailing{
		id:         NewMailingID(),
		botID:      botID,
		name:       name,
		entryKey:   entryKey,
		status:     MailingStatusScheduled,
		recipients: recsMapped,
		results:    make(map[UserID]UserMailingResult),
		createdAt:  time.Now(),
	}

	m.events = append(m.events, MailingScheduled{
		MailingID: m.id,
		Time:      m.createdAt,
	})
	return m, nil
}

func RestoreMailing(
	id MailingID,
	botID BotID,
	name string,
	entryKey EntryKey,
	status MailingStatus,
	recipients []UserID,
	results []UserMailingResult,
	createdAt time.Time,
	startedAt *time.Time,
	completedAt *time.Time,
) (*Mailing, error) {
	if id.IsZero() {
		return nil, errors.New("id is zero")
	}

	if botID.IsZero() {
		return nil, errors.New("bot id is zero")
	}

	if name == "" {
		return nil, errors.New("name is empty")
	}

	if entryKey.IsZero() {
		return nil, errors.New("entry key is zero")
	}

	if status.IsZero() {
		return nil, errors.New("status is zero")
	}

	if len(recipients) == 0 {
		return nil, errors.New("recipients is empty")
	}

	if createdAt.IsZero() {
		return nil, errors.New("createdAt is zero")
	}

	if startedAt != nil && startedAt.IsZero() {
		return nil, errors.New("startedAt is zero")
	}

	if completedAt != nil && completedAt.IsZero() {
		return nil, errors.New("completedAt is zero")
	}

	recsMapped := mapUserIDs(recipients)
	resMapped := mapUserMailingResults(results)

	return &Mailing{
		id:          id,
		botID:       botID,
		name:        name,
		entryKey:    entryKey,
		status:      status,
		recipients:  recsMapped,
		results:     resMapped,
		createdAt:   createdAt,
		startedAt:   startedAt,
		completedAt: completedAt,
	}, nil
}

func mapUserIDs(userIDs []UserID) map[UserID]bool {
	mapped := make(map[UserID]bool)
	for _, user := range userIDs {
		mapped[user] = true
	}
	return mapped
}

func mapUserMailingResults(results []UserMailingResult) map[UserID]UserMailingResult {
	mapped := make(map[UserID]UserMailingResult)
	for _, res := range results {
		mapped[res.UserID()] = res
	}
	return mapped
}

func (m *Mailing) PullEvents() []event.Event {
	ev := m.events
	m.events = nil
	return ev
}

// MarkStarted переводит Mailing из статуса MailingStatusScheduled в
// MailingStatusStarted и устанавливает время начала рассылки. Если текущий
// статус не MailingStatusScheduled, возвращает ошибку
// shared.ErrIllegalStateTransition.
func (m *Mailing) MarkStarted() error {
	if m.status != MailingStatusScheduled {
		return fmt.Errorf("%w: %s -> %s", shared.ErrIllegalStateTransition, m.status, MailingStatusStarted)
	}
	m.status = MailingStatusStarted
	m.startedAt = hlpr.Ptr(time.Now())
	return nil
}

// MarkFailed переводит Mailing в статус MailingStatusFailed. Если текущий статус
// MailingStatusCompleted, возвращает ошибку shared.ErrIllegalStateTransition.
func (m *Mailing) MarkFailed() error {
	if m.status == MailingStatusCompleted {
		return fmt.Errorf("%w: %s -> %s", shared.ErrIllegalStateTransition, m.status, MailingStatusFailed)
	}
	m.status = MailingStatusFailed
	return nil
}

// MessageSent регистрирует результат успешной отправки сообщения получателю.
// Если текущий статус не MailingStatusStarted, возвращает ошибку
// ErrMailingIsNotStarted. Если userID не входит в список получателей, возвращает
// ошибку ErrUserNotInRecipients. Если после регистрации результата отправки
// сообщений всем пользователям рассылки, переводит статус в
// MailingStatusCompleted и устанавливает время завершения рассылки.
func (m *Mailing) MessageSent(userID UserID) error {
	if m.status != MailingStatusStarted {
		return ErrMailingIsNotStarted
	}

	if _, ok := m.recipients[userID]; !ok {
		return fmt.Errorf("%w: %d", ErrUserNotInRecipients, userID)
	}

	res := NewSuccessMailingResult(userID)
	m.results[userID] = res
	if len(m.results) == len(m.recipients) {
		m.status = MailingStatusCompleted
		m.completedAt = hlpr.Ptr(time.Now())
	}
	return nil
}

func (m *Mailing) MessageFailed(userID UserID, msg string) error {
	if m.status != MailingStatusStarted {
		return ErrMailingIsNotStarted
	}

	if _, ok := m.recipients[userID]; !ok {
		return fmt.Errorf("%w: %d", ErrUserNotInRecipients, userID)
	}

	res := NewErrorMailingResult(userID, msg)
	m.results[userID] = res
	if len(m.results) == len(m.recipients) {
		m.status = MailingStatusCompleted
		m.completedAt = hlpr.Ptr(time.Now())
	}
	return nil
}

func (m *Mailing) SuccessCount() int {
	cnt := 0
	for _, res := range m.results {
		if res.Success() {
			cnt++
		}
	}
	return cnt
}

func (m *Mailing) FailCount() int {
	cnt := 0
	for _, res := range m.results {
		if !res.Success() {
			cnt++
		}
	}
	return cnt
}

func (m *Mailing) PendingCount() int {
	return len(m.recipients) - len(m.results)
}

func (m *Mailing) RecipientsTotal() int {
	return len(m.recipients)
}

func (m *Mailing) ID() MailingID {
	return m.id
}

func (m *Mailing) BotID() BotID {
	return m.botID
}

func (m *Mailing) Name() string {
	return m.name
}

func (m *Mailing) EntryKey() EntryKey {
	return m.entryKey
}

func (m *Mailing) Status() MailingStatus {
	return m.status
}

func (m *Mailing) Recipients() []UserID {
	res := make([]UserID, 0, len(m.recipients))
	for userID := range m.recipients {
		res = append(res, userID)
	}
	return res
}

func (m *Mailing) Results() []UserMailingResult {
	res := make([]UserMailingResult, 0, len(m.results))
	for _, r := range m.results {
		res = append(res, r)
	}
	return res
}

func (m *Mailing) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Mailing) StartedAt() *time.Time {
	return m.startedAt
}

func (m *Mailing) CompletedAt() *time.Time {
	return m.completedAt
}
