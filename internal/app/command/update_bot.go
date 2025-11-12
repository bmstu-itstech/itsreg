package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type UpdateBotHandler decorator.CommandHandler[request.UpdateBotCommand]

type updateBotHandler struct {
	br port.BotRepository
}

func (h updateBotHandler) Handle(ctx context.Context, cmd request.UpdateBotCommand) error {
	bot, err := request.BotFromUpdateCommand(cmd)
	if err != nil {
		return err
	}
	return h.br.UpsertBot(ctx, bot)
}

func NewUpdateBotHandler(
	br port.BotRepository,
	l *slog.Logger,
	mc decorator.MetricsClient,
) UpdateBotHandler {
	return decorator.ApplyCommandDecorators(updateBotHandler{br}, l, mc)
}
