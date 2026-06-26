package inmemory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
	"github.com/bmstu-itstech/itsreg/pkg/reqctx"
)

const EventBusCapacity = 256

type EventBus struct {
	mu        sync.RWMutex
	handlers  map[string][]port.EventHandler
	logger    *slog.Logger
	queue     chan job
	done      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewEventBus(l *slog.Logger) (*EventBus, func()) {
	b := &EventBus{
		handlers: make(map[string][]port.EventHandler),
		logger:   l,
		queue:    make(chan job, EventBusCapacity),
		done:     make(chan struct{}),
	}
	b.wg.Add(1)
	go b.worker()
	return b, b.Close
}

func (b *EventBus) Publish(ctx context.Context, events ...event.Event) error {
	for _, ev := range events {
		j := newJob(ev)
		if reqID, ok := reqctx.FromContext(ctx); ok {
			j = j.withRequestID(reqID)
		}
		select {
		case b.queue <- j:
		case <-ctx.Done():
			return ctx.Err()
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

func (b *EventBus) worker() {
	defer b.wg.Done()
	for {
		select {
		case j := <-b.queue:
			ctx := context.Background()
			if j.reqID != "" {
				ctx = reqctx.WithRequestID(ctx, j.reqID)
			}
			b.handle(ctx, j.Event)

		case <-b.done:
			return
		}
	}
}

func (b *EventBus) handle(ctx context.Context, ev event.Event) {
	l := b.logger.With(slog.String("op", "inmemory.EventBus.handle"))

	b.mu.RLock()
	handlers := append([]port.EventHandler(nil), b.handlers[ev.EventName()]...)
	defer b.mu.RUnlock()

	for _, h := range handlers {
		if err := h.Handle(ctx, ev); err != nil {
			l.ErrorContext(ctx, "failed to handle event",
				slog.String("event_name", ev.EventName()),
				slog.String("error", err.Error()),
			)
		}
	}
}

func (b *EventBus) Close() {
	b.logger.DebugContext(context.Background(), "closing event bus")
	b.closeOnce.Do(func() {
		close(b.done)
	})
	b.wg.Wait()
}
