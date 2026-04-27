package bots

import "fmt"

type MailingStatus struct {
	s string
}

var (
	MailingStatusScheduled = MailingStatus{"scheduled"}
	MailingStatusStarted   = MailingStatus{"started"}
	MailingStatusCompleted = MailingStatus{"completed"}
	MailingStatusFailed    = MailingStatus{"failed"}
)

func MailingStatusFromString(s string) (MailingStatus, error) {
	switch s {
	case "scheduled":
		return MailingStatusScheduled, nil
	case "started":
		return MailingStatusStarted, nil
	case "completed":
		return MailingStatusCompleted, nil
	case "failed":
		return MailingStatusFailed, nil
	default:
		return MailingStatus{}, fmt.Errorf("unknown mailing status: %s", s)
	}
}

func (s MailingStatus) IsZero() bool {
	return s.s == ""
}

func (s MailingStatus) String() string {
	return s.s
}
