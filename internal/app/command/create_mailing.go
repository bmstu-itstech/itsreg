package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

type CreateMailingRequest struct {
	ActorID    int64
	BotID      string
	Name       string
	EntryKey   string
	Recipients []int64
}

type CreateMailingResponse struct {
	MailingID string
}

type CreateMailingHandler struct {
	mr  port.MailingRepository
	bmp port.BotMetaProvider
	eb  port.EventBus
	l   *slog.Logger
}

func NewCreateMailingHandler(
	mr port.MailingRepository,
	bmp port.BotMetaProvider,
	eb port.EventBus,
	l *slog.Logger,
) *CreateMailingHandler {
	return &CreateMailingHandler{mr, bmp, eb, l}
}

func (h *CreateMailingHandler) Handle(ctx context.Context, req CreateMailingRequest) (CreateMailingResponse, error) {
	l := h.l.With(
		slog.String("op", "command.CreateMailingHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
	)

	bot, err := h.bmp.BotMeta(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found")
		return CreateMailingResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to find bot by ID")
		return CreateMailingResponse{}, err
	}

	if bot.Deleted {
		l.InfoContext(ctx, "bot deleted")
		return CreateMailingResponse{}, port.ErrBotNotFound
	}

	if req.ActorID != bot.OwnerID {
		l.InfoContext(ctx, "actor cannot create mailing with the bot", slog.Int64("owner_id", bot.OwnerID))
		return CreateMailingResponse{}, port.ErrBotNotFound
	}

	recs := mappers.UserIDsFromDTO(req.Recipients)
	mailing, err := bots.NewMailing(bots.BotID(req.BotID), req.Name, bots.EntryKey(req.EntryKey), recs)
	if errors.As(err, &shared.ValidationError{}) {
		l.InfoContext(ctx, "invalid mailing", slog.String("error", err.Error()))
		return CreateMailingResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to create mailing", slog.String("error", err.Error()))
		return CreateMailingResponse{}, err
	}

	l = l.With(slog.String("mailing_id", mailing.ID().String()))
	err = h.mr.SaveMailing(ctx, mailing)
	if errors.Is(err, port.ErrMailingAlreadyExists) {
		l.WarnContext(ctx, "mailing already exists", slog.String("error", err.Error()))
		return CreateMailingResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to save mailing", slog.String("error", err.Error()))
		return CreateMailingResponse{}, err
	}

	if err = h.eb.Publish(ctx, mailing.PullEvents()...); err != nil {
		l.ErrorContext(ctx, "failed to publish mailing events", slog.String("error", err.Error()))
		return CreateMailingResponse{}, err
	}

	l.InfoContext(ctx, "mailing successfully created")

	return CreateMailingResponse{
		MailingID: mailing.ID().String(),
	}, nil
}
