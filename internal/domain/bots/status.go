package bots

import "fmt"

type Status struct {
	s string
}

var (
	StatusStarting = Status{"starting"}
	StatusActive   = Status{"active"}
	StatusStopped  = Status{"stopped"}
	StatusFailed   = Status{"failed"}
)

func StatusFromString(s string) (Status, error) {
	switch s {
	case "starting":
		return StatusStarting, nil
	case "active":
		return StatusActive, nil
	case "stopped":
		return StatusStopped, nil
	case "failed":
		return StatusFailed, nil
	default:
		return Status{}, fmt.Errorf("unknown status: %s", s)
	}
}

func (s Status) IsZero() bool {
	return s.s == ""
}

func (s Status) String() string {
	return s.s
}
