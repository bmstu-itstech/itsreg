package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/dto/mappers"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type ProcessRequest struct {
	BotID   string
	UserID  int64
	Message dto.Message
}

type ProcessResponse struct{}

type ProcessHandler struct {
	br port.BotRepository
	sr port.ScriptRepository
	tr port.ThreadRepository
	ms port.MessageSender
	l  *slog.Logger
}

func NewProcessHandler(
	br port.BotRepository,
	sr port.ScriptRepository,
	tr port.ThreadRepository,
	ms port.MessageSender,
	l *slog.Logger,
) *ProcessHandler {
	return &ProcessHandler{br, sr, tr, ms, l}
}

func (h *ProcessHandler) Handle(ctx context.Context, req ProcessRequest) (ProcessResponse, error) {
	l := h.l.With(
		slog.String("op", "command.ProcessHandler.Handle"),
		slog.String("bot_id", req.BotID),
		slog.Int64("user_id", req.UserID),
		slog.String("message", req.Message.Text),
	)

	msg, err := mappers.MessageFromDTO(req.Message)
	if err != nil {
		l.ErrorContext(ctx, "failed to create message", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	// Получение агрегатов Bot и Script.
	// В будущем, когда будет происходить разделение bounded contexts, загрузка двух write-моделей будет заменена
	// на чтение read-модели(ей)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	if err = bot.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	l = l.With(slog.String("script_id", bot.ScriptID().String()))

	script, err := h.sr.Script(ctx, bot.ScriptID())
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	if err = script.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	// Process может выполняться только для уже созданного треда

	thread, err := h.tr.LastUserThread(ctx, bot.ID(), bots.UserID(req.UserID))
	if errors.Is(err, port.ErrUserHasNotThreads) {
		l.InfoContext(ctx, "user has not any threads", slog.String("error", err.Error()))
		// Ошибки нет, сообщение игнорируется
		return ProcessResponse{}, nil
	}
	if err != nil {
		l.InfoContext(ctx, "failed to fetch user thread", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	l = l.With(slog.String("thread_id", thread.ID().String()))

	// Непосредственно выполнение обработки сообщения

	res, err := script.Process(thread, msg)
	if err != nil {
		l.InfoContext(ctx, "failed to process message", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}

	// В первую очередь необходимо гарантировать, что изменения в треде будут записаны.

	err = h.tr.UpdateThread(ctx, thread)
	if err != nil {
		l.ErrorContext(ctx, "failed to update thread", slog.String("error", err.Error()))
		return ProcessResponse{}, err
	}
	l.DebugContext(ctx, "thread updated")

	// После внедрения событийной модели, отправка сообщений будет асинхронной
	// через bus.Publish(thread.PullEvents()), что гарантирует их отправку

	for _, m := range res {
		err = h.ms.Send(ctx, bot.Token(), bots.UserID(req.UserID), m)
		if err != nil {
			l.ErrorContext(ctx, "failed to send response message", slog.String("error", err.Error()))
			return ProcessResponse{}, err
		}
	}

	return ProcessResponse{}, nil
}
