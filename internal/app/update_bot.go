package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type UpdateBot struct {
	BotId  string
	Author int64
	Token  string
	Script Script
}

type UpdateBotHandler decorator.CommandHandler[UpdateBot]

type updateBotHandler struct {
	bm bots.BotManager
}

func (h updateBotHandler) Handle(ctx context.Context, cmd UpdateBot) error {
	script, err := scriptFromDto(cmd.Script)
	if err != nil {
		return err
	}

	bot, err := bots.NewBot(bots.BotId(cmd.BotId), bots.Token(cmd.Token), bots.UserId(cmd.Author), script)
	if err != nil {
		return err
	}

	return h.bm.Upsert(ctx, bot)
}

func NewUpdateBotHandler(
	bm bots.BotManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) UpdateBotHandler {
	return decorator.ApplyCommandDecorators(updateBotHandler{bm}, l, mc)
}
