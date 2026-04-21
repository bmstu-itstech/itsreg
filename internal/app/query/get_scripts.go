package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetScriptsRequest struct {
	ActorID int64
}

type GetScriptsResponse = []dto.Script

type GetScriptsHandler struct {
	sr port.ScriptRepository
	l  *slog.Logger
}

func NewGetScriptsHandler(sr port.ScriptRepository, l *slog.Logger) *GetScriptsHandler {
	return &GetScriptsHandler{sr, l}
}

func (h *GetScriptsHandler) Handle(ctx context.Context, req GetScriptsRequest) (GetScriptsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetScriptsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
	)

	scripts, err := h.sr.ScriptsByOwnerID(ctx, bots.UserID(req.ActorID))
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch scripts", slog.String("error", err.Error()))
		return GetScriptsResponse{}, err
	}

	l.InfoContext(ctx, "fetched scripts from repository")

	return mappers.ScriptsToDTO(scripts), nil
}
