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

type GetThreadsTableHandler decorator.QueryHandler[request.GetThreadsTableQuery, response.GetThreadsTableResponse]

type getThreadsTableHandler struct {
	bp  port.BotProvider
	ttp port.ThreadsTableProvider
}

func (h getThreadsTableHandler) Handle(
	ctx context.Context, q request.GetThreadsTableQuery,
) (response.GetThreadsTableResponse, error) {
	_, err := h.bp.Bot(ctx, bots.BotID(q.BotID))
	if err != nil {
		return response.GetThreadsTableResponse{}, err
	}

	//if bot.Author() != bots.UserID(q.Author) {
	//	return response.GetThreadsTableResponse{}, bots.ErrPermissionDenied
	//}

	return h.ttp.ThreadsTable(ctx, bots.BotID(q.BotID))
}

func NewGetThreadsTableHandler(
	bp port.BotProvider,
	ttp port.ThreadsTableProvider,
	l *slog.Logger,
	mc decorator.MetricsClient,
) GetThreadsTableHandler {
	return decorator.ApplyQueryDecorators(getThreadsTableHandler{bp, ttp}, l, mc)
}
