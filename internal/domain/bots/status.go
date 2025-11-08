package bots

type Status struct {
	s string
}

var (
	Running = Status{"running"}
	Idle    = Status{"idle"}
	Dead    = Status{"dead"}
)

func (s *Status) String() string {
	return s.s
}
