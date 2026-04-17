package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type RecoverActiveRunsHandler struct {
	rr port.RunRepository
	eb port.EventBus
	l  *slog.Logger
}

func NewRecoverActiveRunsHandler(rr port.RunRepository, eb port.EventBus, l *slog.Logger) *RecoverActiveRunsHandler {
	return &RecoverActiveRunsHandler{rr, eb, l}
}

func (h *RecoverActiveRunsHandler) Recover(ctx context.Context) error {
	l := h.l.With(slog.String("op", "bootstrap.RecoverActiveRunsHandler.Handle"))

	runs, err := h.rr.ActiveRuns(ctx)
	if err != nil {
		l.ErrorContext(ctx, "failed to fetch active runs", slog.String("error", err.Error()))
		return err
	}

	for _, run := range runs {
		err2 := h.eb.Publish(ctx, bots.RunRecoverRequested{
			RunID: run.ID(),
			Time:  time.Now(),
		})
		if err2 != nil {
			l.WarnContext(ctx, "failed to publish event", slog.String("error", err2.Error()))
			// continue
		}
	}

	return nil
}
