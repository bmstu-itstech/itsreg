package request

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type UpdateBotCommand struct {
	BotID  string
	Author int64
	Token  string
	Script dto.Script
}

func BotFromUpdateCommand(cmd UpdateBotCommand) (bots.Bot, error) {
	script, err := dto.ScriptFromDTO(cmd.Script)
	if err != nil {
		return bots.Bot{}, err
	}
	return bots.NewBot(bots.BotID(cmd.BotID), bots.Token(cmd.Token), bots.UserID(cmd.Author), script)
}
