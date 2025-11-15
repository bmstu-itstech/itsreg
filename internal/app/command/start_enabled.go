package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type StartEnabledHandler decorator.CommandHandler[request.StartEnabledBotsCommand]

type startEnabledHandler struct {
	im port.InstanceManager
	bp port.BotProvider
}

func (h startEnabledHandler) Handle(ctx context.Context, _ request.StartEnabledBotsCommand) error {
	_bots, err := h.bp.EnabledBots(ctx)
	if err != nil {
		return err
	}
	var errs bots.MultiError
	for _, bot := range _bots {
		err = h.im.Start(ctx, bot.ID(), bot.Token())
		if err != nil {
			errs.ExtendOrAppend(err)
		}
	}
	if errs.HasError() {
		return &errs
	}
	return nil
}

func NewStartEnabledHandler(
	im port.InstanceManager,
	bp port.BotProvider,
	l *slog.Logger,
	mc decorator.MetricsClient,
) StartEnabledHandler {
	return decorator.ApplyCommandDecorators(startEnabledHandler{im, bp}, l, mc)
}
