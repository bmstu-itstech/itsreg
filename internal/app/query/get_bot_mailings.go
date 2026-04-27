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

type GetBotMailingsRequest struct {
	ActorID int64
	BotID   string
	Status  *string
}

type GetBotMailingsResponse = []dto.Mailing

type GetBotMailingsHandler struct {
	sr  port.MailingRepository
	bmp port.BotMetaProvider
	l   *slog.Logger
}

func NewGetBotMailingsHandler(
	sr port.MailingRepository, bmp port.BotMetaProvider, l *slog.Logger,
) *GetBotMailingsHandler {
	return &GetBotMailingsHandler{sr, bmp, l}
}

func (h *GetBotMailingsHandler) Handle(ctx context.Context, req GetBotMailingsRequest) (GetBotMailingsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetBotMailingsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
		slog.Any("status", req.Status),
	)

	var filter port.MailingsFilter
	filter.BotID = hlpr.Ptr(bots.BotID(req.BotID))
	if req.Status != nil {
		status, err := bots.MailingStatusFromString(*req.Status)
		if err != nil {
			l.InfoContext(ctx, "invalid status filter", slog.String("error", err.Error()))
			return GetMailingsResponse{}, nil
		}
		filter.Status = &status
	}

	_, err := h.bmp.BotMeta(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found", slog.String("bot_id", req.BotID))
		return GetMailingsResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot from repository", slog.String("error", err.Error()))
		return GetMailingsResponse{}, err
	}

	mailings, err := h.sr.MailingsByOwnerID(ctx, bots.UserID(req.ActorID), filter)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch mailings", slog.String("error", err.Error()))
		return GetMailingsResponse{}, err
	}

	l.InfoContext(ctx, "fetched mailings from repository")

	return mappers.MailingsToDTO(mailings), nil
}
