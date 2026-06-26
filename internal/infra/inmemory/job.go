package inmemory

import "github.com/bmstu-itstech/itsreg/internal/domain/shared/event"

type job struct {
	event.Event

	reqID string
}

func newJob(e event.Event) job {
	return job{
		Event: e,
	}
}

func (j job) withRequestID(id string) job {
	j.reqID = id
	return j
}
