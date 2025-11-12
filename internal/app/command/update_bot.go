package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type UpdateBotHandler decorator.CommandHandler[request.UpdateBotCommand]

type updateBotHandler struct {
	bm bots.BotManager
}

func (h updateBotHandler) Handle(ctx context.Context, cmd request.UpdateBotCommand) error {
	bot, err := request.BotFromUpdateCommand(cmd)
	if err != nil {
		return err
	}
	return h.bm.Upsert(ctx, bot)
}

func NewUpdateBotHandler(
	bm bots.BotManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) UpdateBotHandler {
	return decorator.ApplyCommandDecorators(updateBotHandler{bm}, l, mc)
}
