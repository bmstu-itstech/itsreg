package bots

import "fmt"

type RunStatus struct {
	s string
}

var (
	RunStatusStarting = RunStatus{"starting"}
	RunStatusActive   = RunStatus{"active"}
	RunStatusStopping = RunStatus{"stopping"}
	RunStatusStopped  = RunStatus{"stopped"}
	RunStatusFailed   = RunStatus{"failed"}
)

func RunStatusFromString(s string) (RunStatus, error) {
	switch s {
	case "starting":
		return RunStatusStarting, nil
	case "active":
		return RunStatusActive, nil
	case "stopping":
		return RunStatusStopping, nil
	case "stopped":
		return RunStatusStopped, nil
	case "failed":
		return RunStatusFailed, nil
	default:
		return RunStatus{}, fmt.Errorf("unknown status: %s", s)
	}
}

func (s RunStatus) IsZero() bool {
	return s.s == ""
}

func (s RunStatus) String() string {
	return s.s
}
