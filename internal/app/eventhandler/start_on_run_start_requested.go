package eventhandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type StartOnRunStartRequestedHandler struct {
	rr port.RunRepository
	im port.InstanceManager
	eb port.EventBus
	l  *slog.Logger
}

func NewStartOnRunStartRequestedHandler(
	rr port.RunRepository,
	im port.InstanceManager,
	eb port.EventBus,
	l *slog.Logger,
) *StartOnRunStartRequestedHandler {
	return &StartOnRunStartRequestedHandler{rr, im, eb, l}
}

func (h *StartOnRunStartRequestedHandler) Handle(ctx context.Context, _ev event.Event) error {
	l := h.l.With(
		slog.String("op", "handler.StartOnRunStartRequested.Handle"),
		slog.String("event", _ev.EventName()),
	)

	ev, ok := _ev.(bots.RunStartRequested)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", _ev)
	}

	l = l.With(
		slog.String("run_id", ev.RunID.String()),
		slog.String("bot_id", ev.BotID.String()),
	)

	run, err := h.rr.Run(ctx, ev.RunID)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch run", slog.String("error", err.Error()))
		return err
	}

	l.InfoContext(ctx, "starting bot instance")
	err = h.im.Start(ctx, run.BotID(), run.Token())
	if err != nil {
		l.ErrorContext(ctx, "failed to start bot instance", slog.String("error", err.Error()))

		if err2 := run.Fail(err.Error()); err2 != nil {
			l.ErrorContext(ctx, "failed to mark run as failed", slog.String("error", err2.Error()))
			return err2
		}

		if err2 := h.rr.UpdateRun(ctx, run); err2 != nil {
			l.ErrorContext(ctx, "failed to save run failure", slog.String("error", err2.Error()))
			return err2
		}

		if err2 := h.eb.Publish(ctx, run.PullEvents()...); err2 != nil {
			l.ErrorContext(ctx, "failed to publish events", slog.String("error", err2.Error()))
			return err2
		}

		return nil
	}

	if err = run.Start(); err != nil {
		l.ErrorContext(ctx, "failed to mark run as started", slog.String("error", err.Error()))
		return err
	}

	if err = h.rr.UpdateRun(ctx, run); err != nil {
		l.ErrorContext(ctx, "failed to update run", slog.String("error", err.Error()))
		return err
	}

	if err = h.eb.Publish(ctx, run.PullEvents()...); err != nil {
		l.ErrorContext(ctx, "failed to publish events", slog.String("error", err.Error()))
		return err
	}

	l.InfoContext(ctx, "bot instance started successfully")
	return nil
}
