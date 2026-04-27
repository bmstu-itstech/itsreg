package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetRunsRequest struct {
	ActorID int64
	Status  *string
	BotID   *string
}

type GetRunsResponse = []dto.Run

type GetRunsHandler struct {
	sr port.RunRepository
	l  *slog.Logger
}

func NewGetRunsHandler(sr port.RunRepository, l *slog.Logger) *GetRunsHandler {
	return &GetRunsHandler{sr, l}
}

func (h *GetRunsHandler) Handle(ctx context.Context, req GetRunsRequest) (GetRunsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetRunsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.Any("status", req.Status),
		slog.Any("bot_id", req.BotID),
	)

	var filter port.RunsFilter
	if req.Status != nil {
		status, err := bots.RunStatusFromString(*req.Status)
		if err != nil {
			l.InfoContext(ctx, "invalid status filter", slog.String("error", err.Error()))
			return GetRunsResponse{}, nil
		}
		filter.Status = &status
	}

	if req.BotID != nil {
		botID := bots.BotID(*req.BotID)
		filter.BotID = &botID
	}

	runs, err := h.sr.RunsByOwnerID(ctx, bots.UserID(req.ActorID), filter)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch runs", slog.String("error", err.Error()))
		return GetRunsResponse{}, err
	}

	l.InfoContext(ctx, "fetched runs from repository")

	return mappers.RunsToDTO(runs), nil
}
