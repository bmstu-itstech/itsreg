package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetUserBots struct {
	UserId int64
}

type GetUserBotsHandler decorator.QueryHandler[GetUserBots, []Bot]

type getUserBotsHandler struct {
	bp bots.BotProvider
}

func (h getUserBotsHandler) Handle(ctx context.Context, q GetUserBots) ([]Bot, error) {
	res, err := h.bp.UserBots(ctx, bots.UserId(q.UserId))
	return batchBotToDto(res), err
}

func NewGetUserBotsHandler(bp bots.BotProvider, l *slog.Logger, mc decorator.MetricsClient) GetUserBotsHandler {
	return decorator.ApplyQueryDecorators(getUserBotsHandler{bp}, l, mc)
}
