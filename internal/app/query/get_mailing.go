package query

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type GetMailingRequest struct {
	ActorID   int64
	MailingID string
}

type GetMailingResponse = dto.OwnedMailing

type GetMailingHandler struct {
	orp port.OwnedMailingProvider
	l   *slog.Logger
}

func NewGetMailingHandler(orp port.OwnedMailingProvider, l *slog.Logger) *GetMailingHandler {
	return &GetMailingHandler{orp, l}
}

func (h *GetMailingHandler) Handle(ctx context.Context, req GetMailingRequest) (GetMailingResponse, error) {
	l := h.l.With(
		slog.String("op", "query.GetMailingHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("mailing_id", req.MailingID),
	)

	mailing, err := h.orp.OwnedMailing(ctx, bots.MailingID(req.MailingID))
	if errors.Is(err, port.ErrMailingNotFound) {
		l.InfoContext(ctx, "mailing not found", slog.String("error", err.Error()))
		return GetMailingResponse{}, err
	}
	if err != nil {
		l.InfoContext(ctx, "failed to fetch owned mailing", slog.String("error", err.Error()))
		return GetMailingResponse{}, err
	}

	l.InfoContext(ctx, "fetched owned mailing from repository")

	if mailing.OwnerID != req.ActorID {
		l.InfoContext(ctx, "mailing does not belong to actor", slog.Int64("owner_id", mailing.OwnerID))
		return GetMailingResponse{}, port.ErrMailingNotFound
	}

	return mailing, nil
}
