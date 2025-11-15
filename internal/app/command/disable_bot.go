package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type DisableBotHandler decorator.CommandHandler[request.DisableBotCommand]

type disableBotHandler struct {
	br port.BotRepository
}

func (h disableBotHandler) Handle(ctx context.Context, cmd request.DisableBotCommand) error {
	bot, err := h.br.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}
	bot.Disable()
	return h.br.UpsertBot(ctx, bot)
}

func NewDisableBotHandler(
	bm port.BotRepository,
	l *slog.Logger,
	mc decorator.MetricsClient,
) DisableBotHandler {
	return decorator.ApplyCommandDecorators(disableBotHandler{bm}, l, mc)
}
