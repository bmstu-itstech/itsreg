package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetThreads struct {
	BotId string
}

type GetThreadsHandler decorator.QueryHandler[GetThreads, []Thread]

type getThreadsHandler struct {
	tp bots.ThreadProvider
}

func (h getThreadsHandler) Handle(ctx context.Context, q GetThreads) ([]Thread, error) {
	threads, err := h.tp.BotThreads(ctx, bots.BotId(q.BotId))
	return batchThreadsToDto(threads), err
}

func NewGetThreadsHandler(tp bots.ThreadProvider, l *slog.Logger, mc decorator.MetricsClient) GetThreadsHandler {
	return decorator.ApplyQueryDecorators(getThreadsHandler{tp}, l, mc)
}
