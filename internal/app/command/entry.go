package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type EntryHandler decorator.CommandHandler[request.EntryCommand]

type entryHandler struct {
	bp bots.BotProvider
	pr bots.ParticipantRepository
	ms bots.BotMessageSender
}

func (h entryHandler) Handle(ctx context.Context, cmd request.EntryCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}

	script := bot.Script()
	prtID := bots.NewParticipantID(bots.UserID(cmd.UserID), bots.BotID(cmd.BotID))

	var response []bots.BotMessage
	err = h.pr.UpdateOrCreate(ctx, prtID, func(
		_ context.Context, prt *bots.Participant,
	) error {
		response, err = script.Entry(prt, bots.EntryKey(cmd.Key))
		return err
	})
	if err != nil {
		return err
	}

	for _, msg := range response {
		err = h.ms.Send(ctx, bot.Token(), prtID.UserID(), msg)
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
