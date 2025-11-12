package request

import "github.com/bmstu-itstech/itsreg-bots/internal/app/dto"

type ProcessCommand struct {
	BotID   string
	UserID  int64
	Message dto.Message
}
