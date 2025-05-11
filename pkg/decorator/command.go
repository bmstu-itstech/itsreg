package decorator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

func ApplyCommandDecorators[H any](
	handler CommandHandler[H],
	logger *slog.Logger,
	metricsClient MetricsClient,
) CommandHandler[H] {
	return commandLoggingDecorator[H]{
		base: commandMetricsDecorator[H]{
			base:   handler,
			client: metricsClient,
		},
		logger: logger,
	}
}

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}
