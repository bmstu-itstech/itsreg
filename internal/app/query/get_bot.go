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

type GetBotRequest struct {
	ActorID int64
	BotID   string
}

type GetBotResponse = dto.Bot

type GetBotHandler struct {
	br port.BotRepository
	l  *slog.Logger
}

func NewGetBotHandler(br port.BotRepository, l *slog.Logger) *GetBotHandler {
	return &GetBotHandler{br, l}
}

func (h *GetBotHandler) Handle(ctx context.Context, req GetBotRequest) (GetBotResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetBotHandler.Handle"),
		slog.String("bot_id", req.BotID),
		slog.Int64("actor_id", req.ActorID),
	)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.WarnContext(ctx, "bot not found", slog.String("error", err.Error()))
		return GetBotResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return GetBotResponse{}, err
	}

	l.InfoContext(ctx, "fetched bot from repository")

	if err = bot.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active bot", slog.String("error", err.Error()))
		return GetBotResponse{}, err
	}

	if err = bot.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.WarnContext(ctx, "failed to ensure owned by bot", slog.String("error", err.Error()))
		return GetBotResponse{}, err
	}

	return mappers.BotToDTO(bot), nil
}
