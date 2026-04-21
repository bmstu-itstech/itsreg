package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type EntryRequest struct {
	BotID    string
	UserID   int64
	Username *string
	EntryKey string
}

type EntryResponse struct{}

type EntryHandler struct {
	br port.BotRepository
	sr port.ScriptRepository
	tr port.ThreadRepository
	ur port.UserRepository
	eb port.EventBus
	l  *slog.Logger
}

func NewEntryHandler(
	br port.BotRepository,
	sr port.ScriptRepository,
	tr port.ThreadRepository,
	ur port.UserRepository,
	eb port.EventBus,
	l *slog.Logger,
) *EntryHandler {
	return &EntryHandler{br, sr, tr, ur, eb, l}
}

func (h *EntryHandler) Handle(ctx context.Context, req EntryRequest) (EntryResponse, error) {
	l := h.l.With(
		slog.String("op", "command.EntryHandler.Handle"),
		slog.String("bot_id", req.BotID),
		slog.Int64("user_id", req.UserID),
		slog.String("entry_key", req.EntryKey),
	)

	// Получение агрегатов Bot и Script.
	// В будущем, когда будет происходить разделение bounded contexts, загрузка двух write-моделей будет заменена
	// на чтение read-модели(ей)

	bot, err := h.br.Bot(ctx, bots.BotID(req.BotID))
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	if err = bot.EnsureActive(); err != nil {
		l.WarnContext(ctx, "failed to ensure active script", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	l = l.With(slog.String("script_id", bot.ScriptID().String()))

	script, err := h.sr.Script(ctx, bot.ScriptID())
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

	if req.Username != nil {
		err = h.ur.UpsertUsername(ctx, bots.UserID(req.UserID), bots.Username(*req.Username))
		if err != nil {
			l.ErrorContext(ctx, "failed to upsert username", slog.String("error", err.Error()))
			return EntryResponse{}, err
		}
	}

	// Непосредственно процедура входа в тред

	thread, msgs, err := script.Entry(bot.ID(), bots.UserID(req.UserID), bots.EntryKey(req.EntryKey))
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

	events := make([]event.Event, len(msgs))
	for i, msg := range msgs {
		events[i] = bots.SendMessageRequested{
			BotID:   thread.BotID(),
			UserID:  thread.UserID(),
			Message: msg,
			Time:    time.Now(),
		}
	}
	if err = h.eb.Publish(ctx, events...); err != nil {
		l.ErrorContext(ctx, "failed to publish events", slog.String("error", err.Error()))
		return EntryResponse{}, err
	}

	l.InfoContext(ctx, "entry processed")

	return EntryResponse{}, nil
}
