package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type EntryRequest struct {
	BotID    string
	UserID   int64
	Username string
	EntryKey string
}

type EntryResponse struct{}

type EntryHandler struct {
	br port.BotRepository
	sr port.ScriptRepository
	tr port.ThreadRepository
	ur port.UserRepository
	ms port.MessageSender
	l  *slog.Logger
}

func NewEntryHandler(
	br port.BotRepository,
	sr port.ScriptRepository,
	tr port.ThreadRepository,
	ur port.UserRepository,
	ms port.MessageSender,
	l *slog.Logger,
) *EntryHandler {
	return &EntryHandler{br, sr, tr, ur, ms, l}
}

func (h *EntryHandler) Handle(ctx context.Context, req EntryRequest) (EntryResponse, error) {
	l := h.l.With(
		slog.String("op", "command.EntryHandler.Handle"),
		slog.String("bot_id", req.BotID),
		slog.Int64("user_id", req.UserID),
		slog.String("username", req.Username),
		slog.String("entry_key", req.EntryKey),
	)

	// Получение агрегатов Bot и Script.
	// В будущем, когда будет происходить разделение bounded contexts, загрузка двух write-моделей будет заменена
	// на чтение read-модели(ей)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if errors.Is(err, port.ErrBotNotFound) {
		l.ErrorContext(ctx, "bot not found", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	if err = bot.EnsureActive(); err != nil {
		l.ErrorContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	l = l.With(slog.String("script_id", bot.ScriptID().String()))

	script, err := h.sr.Script(ctx, bot.ScriptID())
	if errors.Is(err, port.ErrScriptNotFound) {
		l.ErrorContext(ctx, "script not found", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	if err = script.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	// Считаем, что пользователь меняет свой никнейм редко, поэтому один раз получаем его Username
	// и запоминаем в БД

	err = h.ur.UpsertUsername(ctx, bots.UserID(req.UserID), bots.Username(req.Username))
	if err != nil {
		l.ErrorContext(ctx, "failed to upsert username", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	// Непосредственно процедура входа в тред

	thread, res, err := script.Entry(bot.ID(), bots.UserID(req.UserID), bots.EntryKey(req.EntryKey))
	if err != nil {
		l.ErrorContext(ctx, "failed to entry in script", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	l = l.With(slog.String("thread_id", thread.ID().String()))

	// В первую очередь необходимо гарантировать, что изменения в треде будут записаны.

	err = h.tr.SaveThread(ctx, thread)
	if errors.Is(err, port.ErrThreadAlreadyExists) {
		l.WarnContext(ctx, "thread already exists", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to save thread", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}
	l.InfoContext(ctx, "thread saved")

	// После внедрения событийной модели, отправка сообщений будет асинхронной
	// через bus.Publish(thread.PullEvents()), что гарантирует их отправку

	for _, msg := range res {
		err = h.ms.Send(ctx, bot.Token(), bots.UserID(req.UserID), msg)
		if err != nil {
			l.ErrorContext(ctx, "failed to send response message", slog.String("error", err.Error()))
			return EntryResponse{}, err
		}
	}

	return EntryResponse{}, nil
}
