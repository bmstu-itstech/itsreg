package bots

type Username string

func (u Username) String() string {
	return string(u)
}
