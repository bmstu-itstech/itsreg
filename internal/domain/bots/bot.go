package bots

import "fmt"

// BotId есть уникальный идентификатор бота.
type BotId string

// Token есть Telegram токен для бота.
type Token string

type Bot struct {
	id     BotId
	token  Token
	author UserId
	script Script
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
		id:     id,
		token:  token,
		script: script,
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
