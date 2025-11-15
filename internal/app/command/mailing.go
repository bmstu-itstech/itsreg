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
	pr port.ParticipantRepository
	ms port.MessageSender
}

func (h mailingHandler) Handle(ctx context.Context, cmd request.MailingCommand) error {
	botID := bots.BotID(cmd.BotID)
	entryKey := bots.EntryKey(cmd.EntryKey)

	bot, err := h.bp.Bot(ctx, botID)
	if err != nil {
		return err
	}

	script := bot.Script()

	var errs bots.MultiError
	for _, user := range cmd.Users {
		prtID := bots.NewParticipantID(bots.UserID(user), botID)

		var response []bots.BotMessage
		err = h.pr.UpdateOrCreateParticipant(ctx, prtID, func(
			_ context.Context, prt *bots.Participant,
		) error {
			response, err = script.Entry(prt, entryKey)
			return err
		})
		if err != nil {
			// Ошибка в операции над участником критична, возвращаем ошибку сразу
			return err
		}

		for _, msg := range response {
			err = h.ms.Send(ctx, bot.Token(), prtID.UserID(), msg)
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
	pr port.ParticipantRepository,
	ms port.MessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) MailingHandler {
	return decorator.ApplyCommandDecorators(mailingHandler{bp, pr, ms}, l, mc)
}
