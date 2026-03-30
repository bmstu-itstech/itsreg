package bots

type UserID int64

func (u UserID) IsZero() bool {
	return u == 0
}

func (u UserID) Int64() int64 {
	return int64(u)
}
