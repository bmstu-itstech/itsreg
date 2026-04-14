package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type CreateBotRequest struct {
	ActorID  int64
	ScriptID string
	Token    string
	Desc     string
}

type CreateBotResponse struct {
	BotID string
}

type CreateBotHandler struct {
	br  port.BotRepository
	smp port.ScriptMetaProvider
	l   *slog.Logger
}

func NewCreateBotHandler(br port.BotRepository, smp port.ScriptMetaProvider, l *slog.Logger) *CreateBotHandler {
	return &CreateBotHandler{br, smp, l}
}

func (h *CreateBotHandler) Handle(ctx context.Context, req CreateBotRequest) (CreateBotResponse, error) {
	l := h.l.With(
		slog.String("op", "command.CreateBotHandler.Handle"),
		slog.Int64("actor_id", req.ActorID),
	)

	script, err := h.smp.ScriptMeta(ctx, bots.ScriptID(req.ScriptID))
	if errors.Is(err, port.ErrScriptNotFound) {
		l.WarnContext(ctx, "script not found",
			slog.String("script_id", req.ScriptID),
			slog.String("error", err.Error()),
		)
		return CreateBotResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to load script meta",
			slog.String("script_id", req.ScriptID),
			slog.String("error", err.Error()),
		)
		return CreateBotResponse{}, err
	}

	if script.Deleted {
		l.WarnContext(ctx, "script already deleted",
			slog.String("script_id", req.ScriptID),
		)
		return CreateBotResponse{}, bots.ErrScriptDeleted
	}

	if req.ActorID != script.OwnerID {
		l.WarnContext(ctx, "actor cannot create bot with the script",
			slog.String("script_id", req.ScriptID),
			slog.Int64("owner_id", script.OwnerID),
		)
		return CreateBotResponse{}, bots.ErrPermissionDenied
	}

	bot, err := bots.NewBot(
		bots.UserID(req.ActorID),
		bots.ScriptID(req.ScriptID),
		bots.Token(req.Token),
		req.Desc,
	)
	if err != nil {
		l.WarnContext(ctx, "failed to create bot", slog.String("error", err.Error()))
		return CreateBotResponse{}, err
	}

	l = l.With(slog.String("bot_id", bot.ID().String()))

	err = h.br.SaveBot(ctx, bot)
	if errors.Is(err, port.ErrBotAlreadyExists) {
		l.WarnContext(ctx, "bot already exists", slog.String("error", err.Error()))
		return CreateBotResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to save bot", slog.String("error", err.Error()))
		return CreateBotResponse{}, err
	}
	l.InfoContext(ctx, "bot saved")

	return CreateBotResponse{
		BotID: bot.ID().String(),
	}, nil
}
