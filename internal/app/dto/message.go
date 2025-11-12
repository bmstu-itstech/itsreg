package dto

import (
	"errors"
	"strconv"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type Message struct {
	Text string
}

type messageBuilder struct {
	state int
}

func (b *messageBuilder) WithState(state int) *messageBuilder {
	b.state = state
	return b
}

func (b *messageBuilder) Build(dto Message) (bots.Message, error) {
	m, err := bots.NewMessage(dto.Text)
	if err != nil {
		return bots.Message{}, b.enrichError(err)
	}
	return m, nil
}

func (b *messageBuilder) BuildAll(dtos []Message) ([]bots.Message, error) {
	var errs bots.MultiError
	res := make([]bots.Message, len(dtos))
	for i, dto := range dtos {
		m, err := b.Build(dto)
		if err != nil {
			errs.Append(err)
		}
		res[i] = m // Просто будет записана пустая структура
	}
	if errs.HasError() {
		return nil, &errs
	}
	return res, nil
}

func (b *messageBuilder) enrichError(err error) error {
	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		if b.state != 0 {
			iiErr.Details["state"] = strconv.Itoa(b.state)
		}
		return iiErr
	}
	return err
}

func MessageFromDTO(dto Message) (bots.Message, error) {
	return bots.NewMessage(dto.Text)
}

func MessageToDTO(m bots.Message) Message {
	return Message{
		Text: m.Text(),
	}
}

func batchMessagesToDTO(messages []bots.Message) []Message {
	res := make([]Message, len(messages))
	for i, m := range messages {
		res[i] = MessageToDTO(m)
	}
	return res
}
