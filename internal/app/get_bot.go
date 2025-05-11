package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetBot struct {
	Id string
}

type GetBotHandler decorator.QueryHandler[GetBot, Bot]

type getBotHandler struct {
	bp bots.BotProvider
}

func (h getBotHandler) Handle(ctx context.Context, q GetBot) (Bot, error) {
	bot, err := h.bp.Bot(ctx, bots.BotId(q.Id))
	return botToDto(bot), err
}

func NewGetBotHandler(bp bots.BotProvider, l *slog.Logger, mc decorator.MetricsClient) GetBotHandler {
	return decorator.ApplyQueryDecorators(getBotHandler{bp}, l, mc)
}
