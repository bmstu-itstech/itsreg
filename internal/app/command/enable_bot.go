package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type EnableBotHandler decorator.CommandHandler[request.EnableBotCommand]

type enableBotHandler struct {
	br port.BotRepository
	im port.InstanceManager
}

func (h enableBotHandler) Handle(ctx context.Context, cmd request.EnableBotCommand) error {
	bot, err := h.br.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}
	bot.Enable()
	err = h.br.UpsertBot(ctx, bot)
	if err != nil {
		return err
	}
	err = h.im.Start(ctx, bots.BotID(cmd.BotID), bot.Token())
	if err != nil {
		// Бот может быть включён в автозапуск, но не запущен. Проведение компенсирующей транзакции
		bot.Disable()
		_ = h.br.UpsertBot(ctx, bot)
	}
	return err
}

func NewEnableBotHandler(
	bm port.BotRepository,
	im port.InstanceManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) EnableBotHandler {
	return decorator.ApplyCommandDecorators(enableBotHandler{bm, im}, l, mc)
}
