package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type EventHandler interface {
	Handle(ctx context.Context, event event.Event) error
}

type EventBus interface {
	Publish(ctx context.Context, events ...event.Event) error
	Subscribe(eventName string, handler EventHandler) error
}
