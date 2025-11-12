package dto

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Bot struct {
	ID     string
	Token  string
	Author int64
	Script Script
}

func BotToDto(bot bots.Bot) Bot {
	return Bot{
		ID:     string(bot.ID()),
		Token:  string(bot.Token()),
		Author: int64(bot.Author()),
		Script: scriptToDTO(bot.Script()),
	}
}

func BatchBotToDto(bs []bots.Bot) []Bot {
	res := make([]Bot, 0, len(bs))
	for _, bot := range bs {
		res = append(res, BotToDto(bot))
	}
	return res
}
