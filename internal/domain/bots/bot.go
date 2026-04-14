package bots

import (
	"errors"
	"fmt"
	"time"
)

const (
	ErrorCodeBotEmptyScriptID ErrorCode = "bot-empty-script-id"
	ErrorCodeBotEmptyToken    ErrorCode = "bot-empty-token"
)

var ErrBotDeleted = errors.New("bot deleted")

type Bot struct {
	id        BotID
	ownerID   UserID
	scriptID  ScriptID
	token     Token
	desc      string
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewBot(ownerID UserID, scriptID ScriptID, token Token, desc string) (*Bot, error) {
	var details []ValidationErrorDetail
	if ownerID.IsZero() {
		// Ошибка не на стороне пользователя
		return nil, errors.New("ownerID is zero")
	}

	if scriptID.IsZero() {
		details = append(
			details, NewValidationErrorDetail("scriptID", ErrorCodeBotEmptyScriptID, "scriptID cannot be zero"),
		)
	}

	if token.IsZero() {
		details = append(details, NewValidationErrorDetail("token", ErrorCodeBotEmptyToken, "token cannot be zero"))
	}

	if len(details) > 0 {
		return nil, NewValidationError(details...)
	}

	return &Bot{
		id:        NewBotID(),
		ownerID:   ownerID,
		scriptID:  scriptID,
		token:     token,
		desc:      desc,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		deletedAt: nil,
	}, nil
}

func MustNewBot(ownerID UserID, scriptID ScriptID, token Token, desc string) *Bot {
	b, err := NewBot(ownerID, scriptID, token, desc)
	if err != nil {
		panic(err)
	}
	return b
}

func RestoreBot(
	id BotID,
	ownerID UserID,
	scriptID ScriptID,
	token Token,
	desc string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Bot, error) {
	if id.IsZero() {
		return nil, errors.New("id is zero")
	}

	if ownerID.IsZero() {
		return nil, errors.New("ownerID is zero")
	}

	if scriptID.IsZero() {
		return nil, errors.New("scriptID is zero")
	}

	if token.IsZero() {
		return nil, errors.New("token is zero")
	}

	if createdAt.IsZero() {
		return nil, errors.New("createdAt is zero")
	}

	if updatedAt.IsZero() {
		return nil, errors.New("updatedAt is zero")
	}

	if deletedAt != nil && deletedAt.IsZero() {
		return nil, errors.New("deletedAt is zero")
	}

	return &Bot{
		id:        id,
		ownerID:   ownerID,
		scriptID:  scriptID,
		token:     token,
		desc:      desc,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}, nil
}

func (b *Bot) EnsureActive() error {
	if b.Deleted() {
		return ErrBotDeleted
	}
	return nil
}

func (b *Bot) EnsureOwnedBy(userID UserID) error {
	if b.ownerID != userID {
		return ErrPermissionDenied
	}
	return nil
}

func (b *Bot) Delete() error {
	if b.Deleted() {
		return fmt.Errorf("cannot delete bot: %w", ErrBotDeleted)
	}
	t := time.Now()
	b.deletedAt = &t
	return nil
}

func (b *Bot) SetScriptID(scriptID ScriptID) error {
	if b.Deleted() {
		return fmt.Errorf("cannot update bot: %w", ErrBotDeleted)
	}
	if scriptID.IsZero() {
		return NewValidationError(
			NewValidationErrorDetail("scriptID", ErrorCodeBotEmptyScriptID, "scriptID cannot be zero"),
		)
	}
	b.scriptID = scriptID
	b.updatedAt = time.Now()
	return nil
}

func (b *Bot) SetToken(token Token) error {
	if b.Deleted() {
		return fmt.Errorf("cannot update bot: %w", ErrBotDeleted)
	}
	if token.IsZero() {
		return NewValidationError(NewValidationErrorDetail("token", ErrorCodeBotEmptyToken, "token cannot be zero"))
	}
	b.token = token
	b.updatedAt = time.Now()
	return nil
}

func (b *Bot) SetDesc(desc string) error {
	if b.Deleted() {
		return fmt.Errorf("cannot update bot: %w", ErrBotDeleted)
	}
	b.desc = desc
	b.updatedAt = time.Now()
	return nil
}

func (b *Bot) ID() BotID {
	return b.id
}

func (b *Bot) OwnerID() UserID {
	return b.ownerID
}

func (b *Bot) ScriptID() ScriptID {
	return b.scriptID
}

func (b *Bot) Token() Token {
	return b.token
}

func (b *Bot) Desc() string {
	return b.desc
}

func (b *Bot) CreatedAt() time.Time {
	return b.createdAt
}

func (b *Bot) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *Bot) DeletedAt() *time.Time {
	return b.deletedAt
}

func (b *Bot) Deleted() bool {
	return b.deletedAt != nil
}
