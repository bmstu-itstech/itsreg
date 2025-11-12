package query

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/response"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetThreadsHandler decorator.QueryHandler[request.GetThreadsQuery, response.GetThreadsResponse]

type getThreadsHandler struct {
	tp port.ThreadProvider
	up port.UsernameProvider
}

func (h getThreadsHandler) Handle(ctx context.Context, q request.GetThreadsQuery) (response.GetThreadsResponse, error) {
	threads, err := h.tp.BotThreads(ctx, bots.BotID(q.BotID))
	if err != nil {
		return nil, err
	}
	res := make([]dto.Thread, len(threads))
	for i, thread := range threads {
		username, err2 := h.up.Username(ctx, thread.UserID())
		if err2 != nil {
			username = bots.Username(fmt.Sprintf("id%d", thread.UserID()))
		}
		res[i] = dto.ThreadToDto(thread.Thread(), string(username))
	}
	return res, nil
}

func NewGetThreadsHandler(
	tp port.ThreadProvider, up port.UsernameProvider, l *slog.Logger, mc decorator.MetricsClient,
) GetThreadsHandler {
	return decorator.ApplyQueryDecorators(getThreadsHandler{tp, up}, l, mc)
}
