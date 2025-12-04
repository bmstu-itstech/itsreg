package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type DeleteBotHandler decorator.CommandHandler[request.DeleteBotCommand]

type deleteBotHandler struct {
	br port.BotRepository
}

func (h deleteBotHandler) Handle(ctx context.Context, command request.DeleteBotCommand) error {
	return h.br.DeleteBot(ctx, bots.BotID(command.BotID))
}

func NewDeleteBotHandler(
	br port.BotRepository,
	l *slog.Logger,
	mc decorator.MetricsClient,
) DeleteBotHandler {
	return decorator.ApplyCommandDecorators(deleteBotHandler{br}, l, mc)
}
