package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type Start struct {
	BotId string
}

type StartHandler decorator.CommandHandler[Start]

type startHandler struct {
	im bots.InstanceManager
	bp bots.BotProvider
}

func (h startHandler) Handle(ctx context.Context, cmd Start) error {
	bot, err := h.bp.Bot(ctx, bots.BotId(cmd.BotId))
	if err != nil {
		return err
	}
	return h.im.Start(ctx, bots.BotId(cmd.BotId), bot.Token())
}

func NewStartHandler(
	im bots.InstanceManager,
	bp bots.BotProvider,
	l *slog.Logger,
	mc decorator.MetricsClient,
) StartHandler {
	return decorator.ApplyCommandDecorators(startHandler{im, bp}, l, mc)
}
