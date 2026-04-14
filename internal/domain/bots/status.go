package bots

import "fmt"

type Status struct {
	s string
}

var (
	Starting = Status{"starting"}
	Active   = Status{"active"}
	Stopped  = Status{"stopped"}
	Failed   = Status{"failed"}
)

func StatusFromString(s string) (Status, error) {
	switch s {
	case "starting":
		return Starting, nil
	case "active":
		return Active, nil
	case "stopped":
		return Stopped, nil
	case "failed":
		return Failed, nil
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
