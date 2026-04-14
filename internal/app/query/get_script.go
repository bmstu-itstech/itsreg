package query

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetScriptRequest struct {
	ActorID  int64
	ScriptID string
}

type GetScriptResponse = dto.Script

type GetScriptHandler struct {
	sr port.ScriptRepository
	l  *slog.Logger
}

func NewGetScriptHandler(sr port.ScriptRepository, l *slog.Logger) *GetScriptHandler {
	return &GetScriptHandler{sr, l}
}

func (h *GetScriptHandler) Handle(ctx context.Context, req GetScriptRequest) (GetScriptResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetScriptHandler.Handle"),
		slog.String("script_id", req.ScriptID),
		slog.Int64("actor_id", req.ActorID),
	)

	script, err := h.sr.Script(ctx, bots.ScriptID(req.ScriptID))
	if errors.Is(err, port.ErrScriptNotFound) {
		l.WarnContext(ctx, "script not found", slog.String("error", err.Error()))
		return GetScriptResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return GetScriptResponse{}, err
	}

	l.InfoContext(ctx, "fetched script from repository")

	if err = script.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return GetScriptResponse{}, port.ErrScriptNotFound
	}

	if err = script.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.WarnContext(ctx, "failed to ensure owned by script", slog.String("error", err.Error()))
		return GetScriptResponse{}, port.ErrScriptNotFound
	}

	return mappers.ScriptToDTO(script), nil
}
