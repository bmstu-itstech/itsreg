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

type CreateScriptRequest struct {
	ActorID int64
	Desc    string
	Nodes   []dto.Node
	Entries []dto.Entry
}

type CreateScriptResponse struct {
	ScriptID string
}

type CreateScriptHandler struct {
	sr port.ScriptRepository
	l  *slog.Logger
}

func NewCreateScriptHandler(sr port.ScriptRepository, l *slog.Logger) *CreateScriptHandler {
	return &CreateScriptHandler{sr, l}
}

func (h *CreateScriptHandler) Handle(ctx context.Context, req CreateScriptRequest) (CreateScriptResponse, error) {
	l := h.l.With(
		slog.String("op", "command.CreateScriptHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
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
		return CreateScriptResponse{}, vErr
	}

	script, err := bots.NewScript(bots.UserID(req.ActorID), req.Desc, nodes, entries)
	if err != nil {
		l.InfoContext(ctx, "failed to create script", slog.String("error", err.Error()))
		return CreateScriptResponse{}, err
	}

	l = l.With(slog.String("script_id", script.ID().String()))

	err = h.sr.SaveScript(ctx, script)
	if errors.Is(err, port.ErrScriptAlreadyExists) {
		l.WarnContext(ctx, "script already exists", slog.String("error", err.Error()))
		return CreateScriptResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to save script", slog.String("error", err.Error()))
		return CreateScriptResponse{}, err
	}

	l.InfoContext(ctx, "script successfully created")

	return CreateScriptResponse{
		ScriptID: script.ID().String(),
	}, nil
}
