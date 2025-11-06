package bots

import (
	"errors"
	"fmt"
	"time"
)

// BotId есть уникальный идентификатор бота.
type BotId string

// Token есть Telegram токен для бота.
type Token string

type Bot struct {
	id        BotId
	token     Token
	author    UserId
	script    Script
	createdAt time.Time
}

func NewBot(id BotId, token Token, author UserId, script Script) (Bot, error) {
	if id == "" {
		return Bot{}, NewInvalidInputError(
			"invalid-bot",
			"expected not empty bot id",
		)
	}

	if token == "" {
		return Bot{}, NewInvalidInputError(
			"invalid-bot",
			"expected not empty bot token",
		)
	}

	if author == 0 {
		return Bot{}, NewInvalidInputError(
			"invalid-bot",
			"expected not empty bot author id",
		)
	}

	if script.IsZero() {
		return Bot{}, fmt.Errorf("empty script")
	}

	return Bot{
		id:        id,
		token:     token,
		author:    author,
		script:    script,
		createdAt: time.Now().Truncate(time.Second),
	}, nil
}

func MustNewBot(id BotId, token Token, author UserId, script Script) Bot {
	b, err := NewBot(id, token, author, script)
	if err != nil {
		panic(err)
	}
	return b
}

func (b Bot) Id() BotId {
	return b.id
}

func (b Bot) Token() Token {
	return b.token
}

func (b Bot) Author() UserId {
	return b.author
}

func (b Bot) Script() Script {
	return b.script
}

func (b Bot) CreatedAt() time.Time {
	return b.createdAt
}

func UnmarshallBot(
	id string,
	token string,
	author int64,
	script Script,
	createdAt time.Time,
) (Bot, error) {
	if id == "" {
		return Bot{}, errors.New("id is empty")
	}

	if token == "" {
		return Bot{}, errors.New("token is empty")
	}

	if author == 0 {
		return Bot{}, errors.New("author id is empty")
	}

	if script.IsZero() {
		return Bot{}, errors.New("script is empty")
	}

	if createdAt.IsZero() {
		return Bot{}, errors.New("createdAt is empty")
	}

	return Bot{
		id:        BotId(id),
		token:     Token(token),
		author:    UserId(author),
		script:    script,
		createdAt: createdAt,
	}, nil
}
