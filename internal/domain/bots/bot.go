package bots

import "fmt"

// BotId есть уникальный идентификатор бота. Совпадает с ником в ТГ (без @).
type BotId string

// Token есть Telegram токен для бота.
type Token string

type Bot struct {
	id     BotId
	token  string
	script Script
}

func NewBot(id BotId, token string, script Script) (Bot, error) {
	if id == "" {
		return Bot{}, NewInvalidInputError(
			"invalid-bot-empty-id",
			"expected not empty bot id",
		)
	}

	if token == "" {
		return Bot{}, NewInvalidInputError(
			"invalid-bot-empty-token",
			"expected not empty bot token",
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

func (b Bot) Id() BotId {
	return b.id
}

func (b Bot) Token() string {
	return b.token
}

func (b Bot) Script() Script {
	return b.script
}
