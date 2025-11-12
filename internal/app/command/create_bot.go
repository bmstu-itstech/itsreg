package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type CreateBotHandler decorator.CommandHandler[request.CreateBotCommand]

type createBotHandler struct {
	br port.BotRepository
}

func (h createBotHandler) Handle(ctx context.Context, cmd request.CreateBotCommand) error {
	bot, err := request.BotFromCommand(cmd)
	if err != nil {
		return err
	}
	return h.br.UpsertBot(ctx, bot)
}

func NewCreateBotHandler(
	bm port.BotRepository,
	l *slog.Logger,
	mc decorator.MetricsClient,
) CreateBotHandler {
	return decorator.ApplyCommandDecorators(createBotHandler{bm}, l, mc)
}
