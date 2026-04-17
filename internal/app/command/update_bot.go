package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type UpdateBotRequest struct {
	ActorID  int64
	BotID    string
	ScriptID *string
	Token    *string
	Desc     *string
}

type UpdateBotResponse = dto.Bot

type UpdateBotHandler struct {
	br  port.BotRepository
	smp port.ScriptMetaProvider
	l   *slog.Logger
}

func NewUpdateBotHandler(br port.BotRepository, smp port.ScriptMetaProvider, l *slog.Logger) *UpdateBotHandler {
	return &UpdateBotHandler{br, smp, l}
}

func (h *UpdateBotHandler) Handle(ctx context.Context, req UpdateBotRequest) (UpdateBotResponse, error) {
	l := h.l.With(
		slog.String("op", "command.UpdateBotHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
		slog.String("bot_id", req.BotID),
	)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.InfoContext(ctx, "bot not found", slog.String("error", err.Error()))
		return UpdateBotResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return UpdateBotResponse{}, err
	}

	if err = bot.EnsureActive(); err != nil {
		l.InfoContext(ctx, "failed to ensure active bot", slog.String("error", err.Error()))
		return UpdateBotResponse{}, port.ErrBotNotFound
	}

	if err = bot.EnsureOwnedBy(bots.UserID(req.ActorID)); err != nil {
		l.InfoContext(ctx, "failed to ensure owned by bot", slog.String("error", err.Error()))
		return UpdateBotResponse{}, port.ErrBotNotFound
	}

	if req.ScriptID != nil {
		scriptID := bots.ScriptID(*req.ScriptID)
		if err = h.updateScriptID(ctx, l, bot, scriptID); err != nil {
			return UpdateBotResponse{}, err
		}
	}

	if req.Token != nil {
		token := bots.Token(*req.Token)
		if err = bot.SetToken(token); err != nil {
			l.InfoContext(ctx, "failed to set token", slog.String("error", err.Error()))
			return UpdateBotResponse{}, err
		}
	}

	if req.Desc != nil {
		desc := *req.Desc
		if err = bot.SetDesc(desc); err != nil {
			l.InfoContext(ctx, "failed to set desc", slog.String("error", err.Error()))
			return UpdateBotResponse{}, err
		}
	}

	err = h.br.UpdateBot(ctx, bot)
	if errors.Is(err, port.ErrTokenAlreadyExists) {
		l.InfoContext(ctx, "token already exists", slog.String("error", err.Error()))
		return UpdateBotResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to update bot", slog.String("error", err.Error()))
		return UpdateBotResponse{}, err
	}

	return mappers.BotToDTO(bot), nil
}

func (h *UpdateBotHandler) updateScriptID(
	ctx context.Context, l *slog.Logger, bot *bots.Bot, scriptID bots.ScriptID,
) error {
	l = l.With(slog.String("script_id", scriptID.String()))

	script, err := h.smp.ScriptMeta(ctx, scriptID)
	if errors.Is(err, port.ErrScriptNotFound) {
		l.InfoContext(ctx, "script not found", slog.String("error", err.Error()))
		return err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to load script meta", slog.String("error", err.Error()))
		return err
	}

	if script.Deleted {
		l.InfoContext(ctx, "script already deleted")
		return port.ErrScriptNotFound
	}

	if bot.OwnerID() != bots.UserID(script.OwnerID) {
		l.InfoContext(ctx, "actor cannot set bot with the script", slog.Int64("owner_id", script.OwnerID))
		return port.ErrScriptNotFound
	}

	if err = bot.SetScriptID(scriptID); err != nil {
		l.InfoContext(ctx, "failed to set scriptID", slog.String("error", err.Error()))
		return err
	}

	return nil
}
