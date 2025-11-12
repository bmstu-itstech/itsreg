package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type StartHandler decorator.CommandHandler[request.StartCommand]

type startHandler struct {
	im port.InstanceManager
	bp port.BotProvider
}

func (h startHandler) Handle(ctx context.Context, cmd request.StartCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}
	return h.im.Start(ctx, bots.BotID(cmd.BotID), bot.Token())
}

func NewStartHandler(
	im port.InstanceManager,
	bp port.BotProvider,
	l *slog.Logger,
	mc decorator.MetricsClient,
) StartHandler {
	return decorator.ApplyCommandDecorators(startHandler{im, bp}, l, mc)
}
