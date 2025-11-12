package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/response"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetUserBotsHandler decorator.QueryHandler[request.GetUserBotsQuery, response.GetUserBotsResponse]

type getUserBotsHandler struct {
	bp bots.BotProvider
}

func (h getUserBotsHandler) Handle(
	ctx context.Context, q request.GetUserBotsQuery,
) (response.GetUserBotsResponse, error) {
	res, err := h.bp.UserBots(ctx, bots.UserID(q.UserID))
	if err != nil {
		return nil, err
	}
	return dto.BatchBotToDto(res), nil
}

func NewGetUserBotsHandler(bp bots.BotProvider, l *slog.Logger, mc decorator.MetricsClient) GetUserBotsHandler {
	return decorator.ApplyQueryDecorators(getUserBotsHandler{bp}, l, mc)
}
