package eventhandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type StopOnRunStopRequestedHandler struct {
	rr port.RunRepository
	im port.InstanceManager
	eb port.EventBus
	l  *slog.Logger
}

func NewStopOnRunStopRequestedHandler(
	rr port.RunRepository,
	im port.InstanceManager,
	eb port.EventBus,
	l *slog.Logger,
) *StopOnRunStopRequestedHandler {
	return &StopOnRunStopRequestedHandler{rr, im, eb, l}
}

func (h *StopOnRunStopRequestedHandler) Handle(ctx context.Context, _ev event.Event) error {
	l := h.l.With(
		slog.String("op", "eventhandler.StopOnRunStopRequstedHandler.Handle"),
		slog.String("event", _ev.EventName()),
	)

	ev, ok := _ev.(bots.RunStopRequested)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", _ev.EventName())
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

	l.InfoContext(ctx, "stopping bot instance")
	err = h.im.Stop(ctx, ev.BotID)
	if err != nil {
		l.ErrorContext(ctx, "failed to stop bot instance", slog.String("error", err.Error()))

		if err2 := run.Failed(err.Error()); err2 != nil {
			l.ErrorContext(ctx, "failed to mark run as failed", slog.String("error", err2.Error()))
			return err2
		}

		if err2 := h.rr.UpdateRun(ctx, run); err2 != nil {
			l.ErrorContext(ctx, "failed to update run", slog.String("error", err2.Error()))
			return err2
		}

		if err2 := h.eb.Publish(ctx, run.PullEvents()...); err2 != nil {
			l.ErrorContext(ctx, "failed to publish events", slog.String("error", err2.Error()))
			return err2
		}

		return err
	}

	if err = run.Stopped(); err != nil {
		l.ErrorContext(ctx, "failed to mark run as stopped", slog.String("error", err.Error()))
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

	return nil
}
