package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetStatus struct {
	BotId string
}

type GetStatusHandler decorator.QueryHandler[GetStatus, string]

type getStatusHandler struct {
	sp bots.StatusProvider
}

func (h getStatusHandler) Handle(ctx context.Context, q GetStatus) (string, error) {
	status, err := h.sp.Status(ctx, bots.BotId(q.BotId))
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func NewGetStatusHandler(sp bots.StatusProvider, l *slog.Logger, mc decorator.MetricsClient) GetStatusHandler {
	return decorator.ApplyQueryDecorators(getStatusHandler{sp}, l, mc)
}
