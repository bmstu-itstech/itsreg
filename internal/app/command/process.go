package command

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type ProcessHandler decorator.CommandHandler[request.ProcessCommand]

type processHandler struct {
	bp port.BotProvider
	pr port.ParticipantRepository
	ms port.MessageSender
}

func (h processHandler) Handle(ctx context.Context, cmd request.ProcessCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}

	script := bot.Script()
	prtID := bots.NewParticipantID(bots.UserID(cmd.UserID), bots.BotID(cmd.BotID))
	message, err := dto.MessageFromDTO(cmd.Message)
	if err != nil {
		return err
	}

	var response []bots.BotMessage
	err = h.pr.UpdateOrCreateParticipant(ctx, prtID, func(
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
	bp port.BotProvider,
	pr port.ParticipantRepository,
	ms port.MessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) ProcessHandler {
	return decorator.ApplyCommandDecorators(processHandler{bp, pr, ms}, l, mc)
}
