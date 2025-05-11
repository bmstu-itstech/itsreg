package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type EntryCommand struct {
	BotId  string
	UserId int64
	Key    string
}

type EntryHandler decorator.CommandHandler[EntryCommand]

type entryHandler struct {
	bp bots.BotProvider
	pr bots.ParticipantRepository
	ms bots.BotMessageSender
}

func (h entryHandler) Handle(ctx context.Context, cmd EntryCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotId(cmd.BotId))
	if err != nil {
		return err
	}

	script := bot.Script()
	prtId := bots.NewParticipantId(bots.UserId(cmd.UserId), bots.BotId(cmd.BotId))

	var response []bots.BotMessage
	err = h.pr.UpdateOrCreate(ctx, prtId, func(
		_ context.Context, prt *bots.Participant,
	) error {
		response, err = script.Entry(prt, bots.EntryKey(cmd.Key))
		return err
	})
	if err != nil {
		return err
	}

	for _, msg := range response {
		err = h.ms.Send(ctx, bot.Token(), prtId.UserId(), msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewEntryHandler(
	bp bots.BotProvider,
	pr bots.ParticipantRepository,
	ms bots.BotMessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) EntryHandler {
	return decorator.ApplyCommandDecorators(entryHandler{bp, pr, ms}, l, mc)
}
