package app

import (
	"context"
	"fmt"
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
	up bots.UsernameProvider
}

func (h getThreadsHandler) Handle(ctx context.Context, q GetThreads) ([]Thread, error) {
	threads, err := h.tp.BotThreads(ctx, bots.BotId(q.BotId))
	if err != nil {
		return nil, err
	}
	res := make([]Thread, len(threads))
	for i, thread := range threads {
		username, err := h.up.Username(ctx, thread.UserId())
		if err != nil {
			username = bots.Username(fmt.Sprintf("id%d", thread.UserId()))
		}
		res[i] = threadToDto(thread.Thread(), string(username))
	}
	return res, nil
}

func NewGetThreadsHandler(tp bots.ThreadProvider, up bots.UsernameProvider, l *slog.Logger, mc decorator.MetricsClient) GetThreadsHandler {
	return decorator.ApplyQueryDecorators(getThreadsHandler{tp, up}, l, mc)
}
