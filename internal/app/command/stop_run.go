package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type StopRunRequest struct {
	ActorID int64
	RunID   string
}

type StopRunResponse struct{}

type StopRunHandler struct {
	rr  port.RunRepository
	bmp port.BotMetaProvider
	eb  port.EventBus
	l   *slog.Logger
}

func NewStopRunHandler(
	rr port.RunRepository,
	bmp port.BotMetaProvider,
	eb port.EventBus,
	l *slog.Logger,
) *StopRunHandler {
	return &StopRunHandler{rr, bmp, eb, l}
}

func (h *StopRunHandler) Handle(ctx context.Context, req StopRunRequest) (StopRunResponse, error) {
	l := h.l.With(
		slog.String("op", "command.StopRunHandle.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("run_id", req.RunID),
	)

	run, err := h.rr.Run(ctx, bots.RunID(req.RunID))
	if errors.Is(err, port.ErrRunNotFound) {
		l.InfoContext(ctx, "run not found", slog.String("error", err.Error()))
		return StopRunResponse{}, port.ErrRunNotFound
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch run from repository", slog.String("error", err.Error()))
		return StopRunResponse{}, err
	}

	bot, err := h.bmp.BotMeta(ctx, run.BotID())
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot meta from repository", slog.String("error", err.Error()))
		return StopRunResponse{}, err
	}

	l = l.With(slog.String("bot_id", bot.ID))

	if bot.Deleted {
		l.InfoContext(ctx, "bot already deleted")
		return StopRunResponse{}, port.ErrRunNotFound
	}

	if bot.OwnerID != req.ActorID {
		l.InfoContext(ctx, "actor cannot stop the run")
		return StopRunResponse{}, port.ErrRunNotFound
	}

	if err = run.Stop(); err != nil {
		l.ErrorContext(ctx, "failed to stop run", slog.String("error", err.Error()))
		return StopRunResponse{}, err
	}

	if err = h.rr.UpdateRun(ctx, run); err != nil {
		l.ErrorContext(ctx, "failed to update run", slog.String("error", err.Error()))
		return StopRunResponse{}, err
	}

	if err = h.eb.Publish(ctx, run.PullEvents()...); err != nil {
		l.ErrorContext(ctx, "failed to publish events", slog.String("error", err.Error()))
		return StopRunResponse{}, err
	}

	l.InfoContext(ctx, "run successfully stopped")

	return StopRunResponse{}, nil
}
