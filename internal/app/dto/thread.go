package dto

import (
	"time"
)

type Thread struct {
	ID        string
	UserID    int64
	Key       string
	StartedAt time.Time
	Username  string
	Answers   map[int]Message
}
