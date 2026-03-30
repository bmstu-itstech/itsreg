package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type EntryHandler decorator.CommandHandler[request.EntryCommand]

type entryHandler struct {
	bp port.BotProvider
	tr port.ThreadRepository
	ur port.UserRepository
	ms port.MessageSender
}

func (h entryHandler) Handle(ctx context.Context, cmd request.EntryCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}

	script := bot.Script()
	userID := bots.UserID(cmd.UserID)

	err = h.ur.UpsertUsername(ctx, bots.UserID(cmd.UserID), bots.Username(cmd.Username))
	if err != nil {
		return err
	}

	var thread *bots.Thread
	var response []bots.BotMessage
	thread, response, err = script.Entry(bot.ID(), userID, bots.EntryKey(cmd.EntryKey))
	if err != nil {
		return err
	}

	err = h.tr.SaveThread(ctx, thread)
	if err != nil {
		return err
	}

	for _, msg := range response {
		err = h.ms.Send(ctx, bot.Token(), userID, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewEntryHandler(
	bp port.BotProvider,
	tr port.ThreadRepository,
	ur port.UserRepository,
	ms port.MessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) EntryHandler {
	return decorator.ApplyCommandDecorators(entryHandler{bp, tr, ur, ms}, l, mc)
}
