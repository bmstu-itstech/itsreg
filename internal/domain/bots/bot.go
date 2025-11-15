package bots

import (
	"errors"
	"time"
)

// BotID есть уникальный идентификатор бота.
type BotID string

// Token есть Telegram токен для бота.
type Token string

type Bot struct {
	id        BotID
	token     Token
	author    UserID
	enabled   bool
	script    Script
	createdAt time.Time
}

func NewBot(id BotID, token Token, author UserID, script Script) (*Bot, error) {
	if id == "" {
		return nil, NewInvalidInputError("bot-empty-id", "expected not empty bot id", "field", "id2")
	}

	if token == "" {
		return nil, NewInvalidInputError("bot-empty-token", "expected not empty bot token", "field", "token")
	}

	if author == 0 {
		return nil, NewInvalidInputError("bot-empty-author-id", "expected not empty bot author", "field", "author")
	}

	if script.IsZero() {
		return nil, errors.New("empty script")
	}

	return &Bot{
		id:        id,
		token:     token,
		author:    author,
		enabled:   false,
		script:    script,
		createdAt: time.Now().Truncate(time.Second),
	}, nil
}

func MustNewBot(id BotID, token Token, author UserID, script Script) *Bot {
	b, err := NewBot(id, token, author, script)
	if err != nil {
		panic(err)
	}
	return b
}

func (b *Bot) Enable() {
	b.enabled = true
}

func (b *Bot) Disable() {
	b.enabled = false
}

func (b *Bot) ID() BotID {
	return b.id
}

func (b *Bot) Token() Token {
	return b.token
}

func (b *Bot) Author() UserID {
	return b.author
}

func (b *Bot) Enabled() bool {
	return b.enabled
}

func (b *Bot) Script() Script {
	return b.script
}

func (b *Bot) CreatedAt() time.Time {
	return b.createdAt
}

func UnmarshallBot(
	id string,
	token string,
	author int64,
	enabled bool,
	script Script,
	createdAt time.Time,
) (*Bot, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	if token == "" {
		return nil, errors.New("token is empty")
	}

	if author == 0 {
		return nil, errors.New("author id is empty")
	}

	if script.IsZero() {
		return nil, errors.New("script is empty")
	}

	if createdAt.IsZero() {
		return nil, errors.New("createdAt is empty")
	}

	return &Bot{
		id:        BotID(id),
		token:     Token(token),
		author:    UserID(author),
		enabled:   enabled,
		script:    script,
		createdAt: createdAt,
	}, nil
}
