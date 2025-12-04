package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type BotRepository interface {
	// UpsertBot создаёт нового бота или обновляет существующий с данным botID.
	UpsertBot(ctx context.Context, bot *bots.Bot) error
	DeleteBot(ctx context.Context, id bots.BotID) error

	BotProvider
}
