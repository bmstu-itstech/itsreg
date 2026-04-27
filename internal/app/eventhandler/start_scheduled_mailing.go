package eventhandler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type StartScheduledMailingHandler struct {
	mr  port.MailingRepository
	tr  port.ThreadRepository
	bmp port.BotMetaProvider
	sr  port.ScriptRepository
	eb  port.EventBus
	l   *slog.Logger
}

func NewStartScheduledMailing(
	mr port.MailingRepository,
	tr port.ThreadRepository,
	bmp port.BotMetaProvider,
	sr port.ScriptRepository,
	eb port.EventBus,
	l *slog.Logger,
) *StartScheduledMailingHandler {
	return &StartScheduledMailingHandler{mr, tr, bmp, sr, eb, l}
}

func (h *StartScheduledMailingHandler) Handle(ctx context.Context, _ev event.Event) error {
	l := h.l.With(
		slog.String("op", "eventhandler.StartScheduledMailingHandler.Handle"),
		slog.String("event", _ev.EventName()),
	)

	ev, ok := _ev.(bots.MailingScheduled)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", _ev)
	}

	l = l.With(slog.String("mailing_id", ev.MailingID.String()))

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

	script, err := h.sr.Script(ctx, bots.ScriptID(bot.ScriptID))
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch script", slog.String("error", err.Error()))
		return err
	}

	err = mailing.MarkStarted()
	if err != nil {
		l.ErrorContext(ctx, "failed to start mailing", slog.String("error", err.Error()))
		return err
	}

	if err = h.mr.UpdateMailing(ctx, mailing); err != nil {
		l.ErrorContext(ctx, "failed to update mailing", slog.String("error", err.Error()))
		return err
	}

	for _, rec := range mailing.Recipients() {
		err = h.handleRecipient(ctx, l, mailing, script, rec)
		if err != nil {
			return err
		}
	}

	l.InfoContext(ctx, "mailing started successfully")

	return nil
}

func (h *StartScheduledMailingHandler) handleRecipient(
	ctx context.Context,
	l *slog.Logger,
	mailing *bots.Mailing,
	script *bots.Script,
	recipient bots.UserID,
) error {
	l = l.With(slog.Int64("recipient_id", recipient.Int64()))

	thread, msgs, err := script.Entry(mailing.BotID(), recipient, mailing.EntryKey())
	if err != nil {
		l.ErrorContext(ctx, "failed to entry in script", slog.String("error", err.Error()))
		return err
	}

	l = l.With(slog.String("thread_id", thread.ID().String()))

	err = h.tr.SaveThread(ctx, thread)
	if err != nil {
		l.ErrorContext(ctx, "failed to save thread", slog.String("error", err.Error()))
		return err
	}

	events := make([]event.Event, len(msgs))
	for i, msg := range msgs {
		events[i] = bots.SendMailingMessageRequested{
			MailingID: mailing.ID(),
			BotID:     thread.BotID(),
			UserID:    thread.UserID(),
			Message:   msg,
			Time:      time.Now(),
		}
	}
	if err = h.eb.Publish(ctx, events...); err != nil {
		l.ErrorContext(ctx, "failed to publish events", slog.String("error", err.Error()))
		return err
	}

	return nil
}
