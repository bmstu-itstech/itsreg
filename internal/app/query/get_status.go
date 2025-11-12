package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/response"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetStatusHandler decorator.QueryHandler[request.GetStatusQuery, response.GetStatusResponse]

type getStatusHandler struct {
	sp port.StatusProvider
}

func (h getStatusHandler) Handle(ctx context.Context, q request.GetStatusQuery) (response.GetStatusResponse, error) {
	status, err := h.sp.Status(ctx, bots.BotID(q.BotID))
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func NewGetStatusHandler(sp port.StatusProvider, l *slog.Logger, mc decorator.MetricsClient) GetStatusHandler {
	return decorator.ApplyQueryDecorators(getStatusHandler{sp}, l, mc)
}
