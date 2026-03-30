package bots

type ThreadID string

func (t ThreadID) IsZero() bool {
	return t == ""
}

func (t ThreadID) String() string {
	return string(t)
}
