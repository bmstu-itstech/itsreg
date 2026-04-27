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

type SendMailingMessageHandler struct {
	ms  port.MessageSender
	mr  port.MailingRepository
	bmp port.BotMetaProvider
	rl  port.RateLimiter
	eb  port.EventBus
	l   *slog.Logger
}

func NewSendMailingMessageHandler(
	ms port.MessageSender,
	mr port.MailingRepository,
	bmp port.BotMetaProvider,
	rl port.RateLimiter,
	eb port.EventBus,
	l *slog.Logger,
) *SendMailingMessageHandler {
	return &SendMailingMessageHandler{ms, mr, bmp, rl, eb, l}
}

func (h *SendMailingMessageHandler) Handle(ctx context.Context, _ev event.Event) error {
	l := h.l.With(
		slog.String("op", "eventhandler.SendMailingMessageHandler.Handle"),
		slog.String("event", _ev.EventName()),
	)

	ev, ok := _ev.(bots.SendMailingMessageRequested)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", _ev)
	}

	l = l.With(slog.Int64("user_id", ev.UserID.Int64()))

	mailing, err := h.mr.Mailing(ctx, ev.MailingID)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch mailing", slog.String("error", err.Error()))
		return err
	}

	bot, err := h.bmp.BotMeta(ctx, mailing.BotID())
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
	if err != nil {
		h.messageFailed(ctx, l, mailing, ev.UserID, err)

		if errors.Is(err, port.ErrUserBlockedBot) {
			l.InfoContext(ctx, "user blocked the bot")
			return nil
		}
		if errors.Is(err, port.ErrMessageSendRateLimitExceeded) {
			l.WarnContext(ctx, "message send rate limit exceeded")
			// В будущем можно настроить переотправку
			return err
		}
		if err != nil {
			l.ErrorContext(ctx, "failed to send mailing message", slog.String("error", err.Error()))
			return err
		}
	}

	if err = mailing.MessageSent(ev.UserID); err != nil {
		l.InfoContext(ctx, "failed to mark mailing message as sent", slog.String("error", err.Error()))
		return err
	}

	if err = h.mr.UpdateMailing(ctx, mailing); err != nil {
		l.ErrorContext(ctx, "failed to update mailing", slog.String("error", err.Error()))
		return err
	}

	l.InfoContext(ctx, "mailing message sent successfully")

	return nil
}

func (h *SendMailingMessageHandler) rateLimitWait(ctx context.Context, token bots.Token) error {
	now := time.Now()
	for {
		wait, err := h.rl.Wait(ctx, token, now)
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
			now = time.Now()
		}
	}
}

func (h *SendMailingMessageHandler) messageFailed(
	ctx context.Context, l *slog.Logger, m *bots.Mailing, userID bots.UserID, err error,
) {
	if err2 := m.MessageFailed(userID, err.Error()); err2 != nil {
		l.ErrorContext(ctx, "failed to mark message sending as failed", slog.String("error", err2.Error()))
	}

	if err2 := h.mr.UpdateMailing(ctx, m); err2 != nil {
		l.ErrorContext(ctx, "failed to update mailing after message sending failure",
			slog.String("error", err2.Error()),
		)
	}
}
