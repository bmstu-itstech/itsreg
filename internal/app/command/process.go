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
	tr port.ThreadRepository
	ms port.MessageSender
}

func (h processHandler) Handle(ctx context.Context, cmd request.ProcessCommand) error {
	bot, err := h.bp.Bot(ctx, bots.BotID(cmd.BotID))
	if err != nil {
		return err
	}

	script := bot.Script()
	message, err := dto.MessageFromDTO(cmd.Message)
	if err != nil {
		return err
	}

	var response []bots.BotMessage

	userID := bots.UserID(cmd.UserID)
	thread, err := h.tr.LastUserThread(ctx, bot.ID(), userID)
	if err != nil {
		return err
	}

	response, err = script.Process(thread, message)
	if err != nil {
		return err
	}

	err = h.tr.UpdateThread(ctx, thread)
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

func NewProcessHandler(
	bp port.BotProvider,
	tr port.ThreadRepository,
	ms port.MessageSender,
	l *slog.Logger,
	mc decorator.MetricsClient,
) ProcessHandler {
	return decorator.ApplyCommandDecorators(processHandler{bp, tr, ms}, l, mc)
}
