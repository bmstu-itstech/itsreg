package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type CreateBot struct {
	BotId  string
	Token  string
	Author int64
	Script Script
}

type CreateBotHandler decorator.CommandHandler[CreateBot]

type createBotHandler struct {
	bm bots.BotManager
}

func (h createBotHandler) Handle(ctx context.Context, cmd CreateBot) error {
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

func NewCreateBotHandler(
	bm bots.BotManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) CreateBotHandler {
	return decorator.ApplyCommandDecorators(createBotHandler{bm}, l, mc)
}
