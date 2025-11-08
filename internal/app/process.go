package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type Process struct {
	BotID   string
	UserID  int64
	Message Message
}

type ProcessHandler decorator.CommandHandler[Process]

type processHandler struct {
	bp bots.BotProvider
	pr bots.ParticipantRepository
	ms bots.BotMessageSender
}

func (h processHandler) Handle(ctx context.Context, cmd Process) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}

	script := bot.Script()
	prtID := bots.NewParticipantID(bots.UserID(cmd.UserID), bots.BotID(cmd.BotID))
	message, err := messageFromDto(cmd.Message)
	if err != nil {
		return err
	}

	var response []bots.BotMessage
	err = h.pr.UpdateOrCreate(ctx, prtID, func(
		_ context.Context, prt *bots.Participant,
	) error {
		response, err = script.Process(prt, message)
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

func NewProcessHandler(
	bp bots.BotProvider,
	pr bots.ParticipantRepository,
	ms bots.BotMessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) ProcessHandler {
	return decorator.ApplyCommandDecorators(processHandler{bp, pr, ms}, l, mc)
}
