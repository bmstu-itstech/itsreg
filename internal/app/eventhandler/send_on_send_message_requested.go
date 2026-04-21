package eventhandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type SendOnSendMessageRequestedHandler struct {
	ms  port.MessageSender
	bmp port.BotMetaProvider
	rl  port.RateLimiter
	eb  port.EventBus
	l   *slog.Logger
}

func NewSendOnSendMessageRequestedHandler(
	ms port.MessageSender,
	bmp port.BotMetaProvider,
	rl port.RateLimiter,
	eb port.EventBus,
	l *slog.Logger,
) *SendOnSendMessageRequestedHandler {
	return &SendOnSendMessageRequestedHandler{ms, bmp, rl, eb, l}
}

func (h *SendOnSendMessageRequestedHandler) Handle(ctx context.Context, _ev event.Event) error {
	l := h.l.With(
		slog.String("op", "eventhandler.SendOnSendMessageRequestedHandler.Handle"),
		slog.String("event", _ev.EventName()),
	)

	ev, ok := _ev.(bots.SendMessageRequested)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", _ev)
	}

	l = l.With(slog.Int64("user_id", ev.UserID.Int64()))

	bot, err := h.bmp.BotMeta(ctx, ev.BotID)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch bot", slog.String("error", err.Error()))
		return err
	}
	token := bots.Token(bot.Token)

	err = h.rateLimitWait(ctx, token)
	if err != nil {
		l.ErrorContext(ctx, "rate limit wait error", slog.String("error", err.Error()))
		return err
	}

	err = h.ms.Send(ctx, token, ev.UserID, ev.Message)
	if errors.Is(err, port.ErrUserBlockedBot) {
		l.InfoContext(ctx, "user blocked the bot, skipping")
		return nil
	}
	if errors.Is(err, port.ErrMessageSendRateLimitExceeded) {
		l.WarnContext(ctx, "message send rate limit exceeded")
		// В будущем можно настроить переотправку
		return err
	}
	if err != nil {
		l.ErrorContext(ctx, "failed to send message")
		return err
	}

	l.InfoContext(ctx, "message sent successfully")

	return nil
}

func (h *SendOnSendMessageRequestedHandler) rateLimitWait(ctx context.Context, token bots.Token) error {
	for {
		wait, err := h.rl.Wait(ctx, token, time.Now())
		if err != nil {
			return err
		}

		if wait == 0 {
			return nil
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
}
