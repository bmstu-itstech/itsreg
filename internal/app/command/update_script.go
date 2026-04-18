package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

type UpdateScriptRequest struct {
	ActorID  int64
	ScriptID string
	Desc     string
	Entries  []dto.Entry
	Nodes    []dto.Node
}

type UpdateScriptResponse = dto.Script

type UpdateScriptHandler struct {
	sr port.ScriptRepository
	l  *slog.Logger
}

func NewUpdateScriptHandler(sr port.ScriptRepository, l *slog.Logger) *UpdateScriptHandler {
	return &UpdateScriptHandler{sr, l}
}

func (h *UpdateScriptHandler) Handle(ctx context.Context, req UpdateScriptRequest) (UpdateScriptResponse, error) {
	l := h.l.With(
		slog.String("op", "command.UpdateScriptHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("script_id", req.ScriptID),
	)

	var vErr shared.ValidationError

	nodes, err := mappers.NodesFromDTOPrefixed(req.Nodes, "nodes")
	if err != nil {
		l.InfoContext(ctx, "failed to map nodes from DTO", slog.String("error", err.Error()))
		vErr = vErr.AppendError(err)
	}

	entries, err := mappers.EntriesFromDTOPrefixed(req.Entries, "entries")
	if err != nil {
		l.InfoContext(ctx, "failed to map entries from DTO", slog.String("error", err.Error()))
		vErr = vErr.AppendError(err)
	}

	if vErr.OrError() != nil {
		return UpdateScriptResponse{}, vErr
	}

	script, err := h.sr.Script(ctx, bots.ScriptID(req.ScriptID))
	if errors.Is(err, port.ErrScriptNotFound) {
		l.InfoContext(ctx, "script not found", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, err
	}

	if err = script.EnsureActive(); err != nil {
		l.InfoContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, port.ErrScriptNotFound
	}

	if err = script.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.InfoContext(ctx, "failed to ensure owned by bot", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, port.ErrScriptNotFound
	}

	err = script.Replace(req.Desc, nodes, entries)
	if errors.As(err, &vErr) {
		l.InfoContext(ctx, "invalid script", slog.String("error", vErr.Error()))
		return UpdateScriptResponse{}, vErr
	}
	if err != nil {
		l.InfoContext(ctx, "failed to replace script", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, err
	}

	if err = h.sr.UpdateScript(ctx, script); err != nil {
		l.ErrorContext(ctx, "failed to update script", slog.String("error", err.Error()))
		return UpdateScriptResponse{}, err
	}

	l.InfoContext(ctx, "script successfully updated")

	return mappers.ScriptToDTO(script), nil
}
