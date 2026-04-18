package query

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetRunRequest struct {
	ActorID int64
	RunID   string
}

type GetRunResponse = dto.OwnedRun

type GetRunHandler struct {
	orp port.OwnedRunProvider
	l   *slog.Logger
}

func NewGetRunHandler(orp port.OwnedRunProvider, l *slog.Logger) *GetRunHandler {
	return &GetRunHandler{orp, l}
}

func (h *GetRunHandler) Handle(ctx context.Context, req GetRunRequest) (GetRunResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetRunHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("run_id", req.RunID),
	)

	run, err := h.orp.OwnedRun(ctx, bots.RunID(req.RunID))
	if errors.Is(err, port.ErrRunNotFound) {
		l.InfoContext(ctx, "run not found", slog.String("error", err.Error()))
		return GetRunResponse{}, err
	}
	if err != nil {
		l.InfoContext(ctx, "failed to fetch owned run", slog.String("error", err.Error()))
		return GetRunResponse{}, err
	}

	if run.OwnerID != req.ActorID {
		l.InfoContext(ctx, "run does not belong to actor", slog.Int64("owner_id", run.OwnerID))
		return GetRunResponse{}, port.ErrRunNotFound
	}

	return run, nil
}
