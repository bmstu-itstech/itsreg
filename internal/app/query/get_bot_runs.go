package query

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
)

type GetBotRunsRequest struct {
	ActorID int64
	BotID   string
	Status  *string
}

type GetBotRunsResponse = []dto.Run

type GetBotRunsHandler struct {
	sr  port.RunRepository
	bmp port.BotMetaProvider
	l   *slog.Logger
}

func NewGetBotRunsHandler(sr port.RunRepository, bmp port.BotMetaProvider, l *slog.Logger) *GetBotRunsHandler {
	return &GetBotRunsHandler{sr, bmp, l}
}

func (h *GetBotRunsHandler) Handle(ctx context.Context, req GetBotRunsRequest) (GetBotRunsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetBotRunsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
		slog.Any("status", req.Status),
	)

	var filter port.RunsFilter
	filter.BotID = hlpr.Ptr(bots.BotID(req.BotID))
	if req.Status != nil {
		status, err := bots.StatusFromString(*req.Status)
		if err != nil {
			l.InfoContext(ctx, "invalid status filter", slog.String("error", err.Error()))
			return GetRunsResponse{}, nil
		}
		filter.Status = &status
	}

	_, err := h.bmp.BotMeta(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found", slog.String("bot_id", req.BotID))
		return GetRunsResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot from repository", slog.String("error", err.Error()))
		return GetRunsResponse{}, err
	}

	runs, err := h.sr.RunsByOwnerID(ctx, bots.UserID(req.ActorID), filter)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch runs", slog.String("error", err.Error()))
		return GetRunsResponse{}, err
	}

	l.InfoContext(ctx, "fetched runs from repository")

	return mappers.RunsToDTO(runs), nil
}
