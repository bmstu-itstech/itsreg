package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type DeleteScriptRequest struct {
	ActorID  int64
	ScriptID string
}

type DeleteScriptResponse struct{}

type DeleteScriptHandler struct {
	sr port.ScriptRepository
	l  *slog.Logger
}

func NewDeleteScriptHandler(sr port.ScriptRepository, l *slog.Logger) *DeleteScriptHandler {
	return &DeleteScriptHandler{
		sr: sr,
		l:  l,
	}
}

func (h *DeleteScriptHandler) Handle(ctx context.Context, req DeleteScriptRequest) (DeleteScriptResponse, error) {
	l := h.l.With(
		slog.String("op", "command.DeleteScriptHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("script_id", req.ScriptID),
	)

	script, err := h.sr.Script(ctx, bots.ScriptID(req.ScriptID))
	if errors.Is(err, port.ErrScriptNotFound) {
		l.WarnContext(ctx, "script not found", slog.String("error", err.Error()))
		return DeleteScriptResponse{}, nil
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return DeleteScriptResponse{}, err
	}

	l.InfoContext(ctx, "fetched script from repository")

	if err = script.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		l.WarnContext(ctx, "script already deleted")
		return DeleteScriptResponse{}, nil
	}

	if err = script.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.WarnContext(ctx, "failed to ensure owned by script", slog.String("error", err.Error()))
		return DeleteScriptResponse{}, port.ErrScriptNotFound
	}

	if err = script.Delete(); err != nil {
		l.ErrorContext(ctx, "failed to delete script", slog.String("error", err.Error()))
		return DeleteScriptResponse{}, err
	}

	if err = h.sr.UpdateScript(ctx, script); err != nil {
		l.ErrorContext(ctx, "failed to update script", slog.String("error", err.Error()))
		return DeleteScriptResponse{}, err
	}

	return DeleteScriptResponse{}, nil
}
