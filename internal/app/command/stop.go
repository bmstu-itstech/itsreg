package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type StopHandler decorator.CommandHandler[request.StopCommand]

type stopHandler struct {
	im bots.InstanceManager
}

func (h stopHandler) Handle(ctx context.Context, cmd request.StopCommand) error {
	return h.im.Stop(ctx, bots.BotID(cmd.BotID))
}

func NewStopHandler(
	im bots.InstanceManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) StopHandler {
	return decorator.ApplyCommandDecorators(stopHandler{im}, l, mc)
}
