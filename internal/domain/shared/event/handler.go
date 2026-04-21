package event

import "context"

type Handler[T Event] interface {
	Handle(ctx context.Context, event T) error
}
