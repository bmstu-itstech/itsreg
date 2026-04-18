package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type CreateRunRequest struct {
	ActorID int64
	BotID   string
}

type CreateRunResponse struct {
	RunID string
}

type CreateRunHandler struct {
	rr  port.RunRepository
	bmp port.BotMetaProvider
	eb  port.EventBus
	l   *slog.Logger
}

func NewCreateRunHandler(
	rr port.RunRepository,
	bmp port.BotMetaProvider,
	eb port.EventBus,
	l *slog.Logger,
) *CreateRunHandler {
	return &CreateRunHandler{rr, bmp, eb, l}
}

func (h *CreateRunHandler) Handle(ctx context.Context, req CreateRunRequest) (CreateRunResponse, error) {
	l := h.l.With(
		slog.String("op", "command.CreateRunHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
	)

	bot, err := h.bmp.BotMeta(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found")
		return CreateRunResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to find bot by ID")
		return CreateRunResponse{}, err
	}

	if bot.Deleted {
		l.InfoContext(ctx, "bot deleted")
		return CreateRunResponse{}, port.ErrBotNotFound
	}

	if req.ActorID != bot.OwnerID {
		l.InfoContext(ctx, "actor cannot create run with the bot", slog.Int64("owner_id", bot.OwnerID))
		return CreateRunResponse{}, port.ErrBotNotFound
	}

	run, err := bots.NewRun(bots.BotID(req.BotID), bots.Token(bot.Token))
	if err != nil {
		l.ErrorContext(ctx, "failed to create run", slog.String("error", err.Error()))
		return CreateRunResponse{}, err
	}

	l = l.With(slog.String("run_id", run.ID().String()))

	err = h.rr.SaveRun(ctx, run)
	if errors.Is(err, port.ErrActiveRunAlreadyExists) {
		l.WarnContext(ctx, "run already exists", slog.String("error", err.Error()))
		return CreateRunResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to save run", slog.String("error", err.Error()))
		return CreateRunResponse{}, err
	}

	if err = h.eb.Publish(ctx, run.PullEvents()...); err != nil {
		l.ErrorContext(ctx, "failed to publish events", slog.String("error", err.Error()))
		return CreateRunResponse{}, err
	}

	l.InfoContext(ctx, "run successfully created")

	return CreateRunResponse{
		RunID: run.ID().String(),
	}, nil
}
