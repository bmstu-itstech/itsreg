package inmemory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
	"github.com/bmstu-itstech/itsreg/pkg/reqctx"
)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]port.EventHandler
	logger   *slog.Logger
}

func NewEventBus(l *slog.Logger) *EventBus {
	return &EventBus{
		handlers: make(map[string][]port.EventHandler),
		logger:   l,
	}
}

func (b *EventBus) Publish(ctx context.Context, events ...event.Event) error {
	l := b.logger.With(slog.String("op", "inmemory.EventBus.Publish"))

	for _, ev := range events {
		b.mu.RLock()
		handlers := b.handlers[ev.EventName()]
		b.mu.RUnlock()

		for _, h := range handlers {
			go func(ctx context.Context, h port.EventHandler, ev event.Event) {
				innerCtx := context.Background()
				if reqID, ok := reqctx.FromContext(ctx); ok {
					innerCtx = reqctx.WithRequestID(innerCtx, reqID)
				}
				if err := h.Handle(innerCtx, ev); err != nil {
					l.ErrorContext(innerCtx, "failed to handle event",
						slog.String("event_name", ev.EventName()),
						slog.String("error", err.Error()),
					)
				}
			}(ctx, h, ev)
		}
	}
	return nil
}

func (b *EventBus) Subscribe(eventName string, handler port.EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	return nil
}
