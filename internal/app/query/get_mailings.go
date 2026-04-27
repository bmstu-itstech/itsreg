package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetMailingsRequest struct {
	ActorID int64
	Status  *string
	BotID   *string
}

type GetMailingsResponse = []dto.Mailing

type GetMailingsHandler struct {
	sr port.MailingRepository
	l  *slog.Logger
}

func NewGetMailingsHandler(sr port.MailingRepository, l *slog.Logger) *GetMailingsHandler {
	return &GetMailingsHandler{sr, l}
}

func (h *GetMailingsHandler) Handle(ctx context.Context, req GetMailingsRequest) (GetMailingsResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetMailingsHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.Any("status", req.Status),
		slog.Any("bot_id", req.BotID),
	)

	var filter port.MailingsFilter
	if req.Status != nil {
		status, err := bots.MailingStatusFromString(*req.Status)
		if err != nil {
			l.InfoContext(ctx, "invalid status filter", slog.String("error", err.Error()))
			return GetMailingsResponse{}, nil
		}
		filter.Status = &status
	}

	if req.BotID != nil {
		botID := bots.BotID(*req.BotID)
		filter.BotID = &botID
	}

	mailings, err := h.sr.MailingsByOwnerID(ctx, bots.UserID(req.ActorID), filter)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch mailings", slog.String("error", err.Error()))
		return GetMailingsResponse{}, err
	}

	l.InfoContext(ctx, "fetched mailings from repository")

	return mappers.MailingsToDTO(mailings), nil
}
