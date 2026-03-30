package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type MailingHandler decorator.CommandHandler[request.MailingCommand]

type mailingHandler struct {
	bp port.BotProvider
	tr port.ThreadRepository
	ms port.MessageSender
}

func (h mailingHandler) Handle(ctx context.Context, cmd request.MailingCommand) error {
	botID := bots.BotID(cmd.BotID)

	bot, err := h.bp.Bot(ctx, botID)
	if err != nil {
		return err
	}

	script := bot.Script()

	var errs bots.MultiError
	for _, user := range cmd.Users {
		userID := bots.UserID(user)

		var thread *bots.Thread
		var response []bots.BotMessage
		thread, response, err = script.Entry(botID, userID, bots.EntryKey(cmd.EntryKey))
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

		for _, msg := range response {
			err = h.ms.Send(ctx, bot.Token(), userID, msg)
			if err != nil {
				// Если ошибка отправки конкретному пользователю, это не должно повлиять на ход рассылки
				errs.Append(err)
			}
		}
	}

	if errs.HasError() {
		return &errs
	}
	return nil
}

func NewMailingHandler(
	bp port.BotProvider,
	tr port.ThreadRepository,
	ms port.MessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) MailingHandler {
	return decorator.ApplyCommandDecorators(mailingHandler{bp, tr, ms}, l, mc)
}
