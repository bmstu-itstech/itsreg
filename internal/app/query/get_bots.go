package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetBotsRequest struct {
	ActorID int64
}

type GetBotsResponse = []dto.Bot

type GetBotsHandler struct {
	br port.BotRepository
	l  *slog.Logger
}

func NewGetBotsHandler(br port.BotRepository, l *slog.Logger) *GetBotsHandler {
	return &GetBotsHandler{br, l}
}

func (h *GetBotsHandler) Handle(ctx context.Context, req GetBotsRequest) (GetBotsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetBotsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
	)

	bs, err := h.br.BotsByOwnerID(ctx, bots.UserID(req.ActorID))
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bots", slog.String("error", err.Error()))
		return GetBotsResponse{}, err
	}

	l.InfoContext(ctx, "fetched bots from repository")

	return mappers.BotsToDTO(bs), nil
}
