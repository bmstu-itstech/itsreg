package event

import "time"

type Event interface {
	EventName() string
	OccurredAt() time.Time
}
