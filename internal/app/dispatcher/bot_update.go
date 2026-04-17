package dispatcher

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
)

type InboundDispatcher struct {
	ph *command.ProcessHandler
	eh *command.EntryHandler
	l  *slog.Logger
}

func NewInboundDispatcher(l *slog.Logger) *InboundDispatcher {
	return &InboundDispatcher{l: l}
}

func (d *InboundDispatcher) SetProcessHandler(ph *command.ProcessHandler) {
	d.ph = ph
}

func (d *InboundDispatcher) SetEntryHandler(eh *command.EntryHandler) {
	d.eh = eh
}

func (d *InboundDispatcher) Dispatch(ctx context.Context, in dto.InboundMessage) error {
	l := d.l.With(
		slog.String("op", "dispatcher.InboundDispatcher.Dispatch"),
		slog.String("bot_id", in.BotID),
		slog.Int64("user_id", in.UserID),
	)

	if in.Text == "" {
		l.InfoContext(ctx, "empty message received; skipping")
		return nil
	}

	if in.IsCommand && d.eh != nil {
		_, err := d.eh.Handle(ctx, command.EntryRequest{
			BotID:    in.BotID,
			UserID:   in.UserID,
			Username: in.Username,
			EntryKey: in.Text,
		})
		return err
	}

	if d.ph != nil {
		_, err := d.ph.Handle(ctx, command.ProcessRequest{
			BotID:  in.BotID,
			UserID: in.UserID,
			Message: dto.Message{
				Text: in.Text,
			},
		})
		return err
	}

	return nil
}
