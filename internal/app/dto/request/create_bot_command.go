package request

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type CreateBotCommand struct {
	BotID  string
	Token  string
	Author int64
	Script dto.Script
}

func BotFromCommand(cmd CreateBotCommand) (bots.Bot, error) {
	script, err := dto.ScriptFromDTO(cmd.Script)
	if err != nil {
		return bots.Bot{}, err
	}
	return bots.NewBot(bots.BotID(cmd.BotID), bots.Token(cmd.Token), bots.UserID(cmd.Author), script)
}
