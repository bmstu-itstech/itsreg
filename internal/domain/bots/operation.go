package bots

// Operation описывает действие, которое будет произведено над Participant
// после обработки Message.
type Operation interface {
	Apply(thr *Thread, in Message)
}

// NoOp не производит никаких действий над пользователем.
type NoOp struct{}

func (a NoOp) Apply(_ *Thread, _ Message) {}

// SaveOp вызывает для пользователя Participant.SaveAnswer.
type SaveOp struct{}

func (a SaveOp) Apply(thr *Thread, in Message) {
	thr.SaveAnswer(in)
}

// AppendOp вызывает для пользователя Participant.AppendAnswer.
type AppendOp struct{}

func (a AppendOp) Apply(thr *Thread, in Message) {
	thr.AppendAnswer(in)
}
