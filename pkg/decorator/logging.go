package decorator

import (
	"context"
	"fmt"
	"log/slog"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger *slog.Logger
}

//nolint:nonamedreturns // так как err используется в defer
func (d commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	logger := d.logger.With(
		slog.String("command", handlerType),
		slog.String("command_body", fmt.Sprintf("%v", cmd)),
	)

	logger.DebugContext(ctx, "executing command")
	defer func() {
		if err == nil {
			logger.InfoContext(ctx, "command executed successfully")
		} else {
			logger.ErrorContext(ctx, "failed to execute command", "error", err.Error())
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type queryLoggingDecorator[C any, R any] struct {
	base   QueryHandler[C, R]
	logger *slog.Logger
}

//nolint:nonamedreturns // так как err используется в defer
func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (_ R, err error) {
	logger := d.logger.With(
		slog.String("query", generateActionName(cmd)),
		slog.String("query_body", fmt.Sprintf("%v", cmd)),
	)

	logger.DebugContext(ctx, "executing query")
	defer func() {
		if err == nil {
			logger.InfoContext(ctx, "query executed successfully")
		} else {
			logger.ErrorContext(ctx, "failed to execute query", "error", err.Error())
		}
	}()

	return d.base.Handle(ctx, cmd)
}
