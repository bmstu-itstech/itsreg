package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
)

type DeleteBotRequest struct {
	ActorID int64
	BotID   string
}

type DeleteBotResponse struct{}

type DeleteBotHandler struct {
	br port.BotRepository
	rr port.RunRepository
	l  *slog.Logger
}

func NewDeleteBotHandler(br port.BotRepository, rr port.RunRepository, l *slog.Logger) *DeleteBotHandler {
	return &DeleteBotHandler{br, rr, l}
}

func (h *DeleteBotHandler) Handle(ctx context.Context, req DeleteBotRequest) (DeleteBotResponse, error) {
	l := h.l.With(
		slog.String("op", "command.DeleteBotHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
	)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found", slog.String("error", err.Error()))
		return DeleteBotResponse{}, nil
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return DeleteBotResponse{}, err
	}

	if err = bot.EnsureActive(); err != nil {
		l.InfoContext(ctx, "bot already deleted")
		return DeleteBotResponse{}, nil
	}

	if err = bot.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.InfoContext(ctx, "failed to ensure owned by bot", slog.String("error", err.Error()))
		return DeleteBotResponse{}, port.ErrBotNotFound
	}

	f := port.RunsFilter{
		BotID:  hlpr.Ptr(bots.BotID(req.BotID)),
		Status: &bots.RunStatusActive,
	}
	runs, err := h.rr.RunsByOwnerID(ctx, bots.UserID(req.ActorID), f)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch runs", slog.String("error", err.Error()))
		return DeleteBotResponse{}, err
	}

	if len(runs) > 0 {
		l.InfoContext(ctx, "can't delete bot with active runs", slog.String("run_id", runs[0].ID().String()))
		return DeleteBotResponse{}, shared.ErrBotHasActiveRuns
	}

	if err = bot.Delete(); err != nil {
		l.ErrorContext(ctx, "failed to delete bot", slog.String("error", err.Error()))
		return DeleteBotResponse{}, err
	}

	if err = h.br.UpdateBot(ctx, bot); err != nil {
		l.ErrorContext(ctx, "failed to update bot", slog.String("error", err.Error()))
		return DeleteBotResponse{}, err
	}

	l.InfoContext(ctx, "bot successfully deleted")

	return DeleteBotResponse{}, nil
}
