package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

func MessageFromDTO(d dto.Message) (bots.Message, error) {
	return bots.NewMessage(d.Text)
}

func messagesFromDTOPrefixed(ds []dto.Message, prefix string) ([]bots.Message, error) {
	var vErr shared.ValidationError
	res := make([]bots.Message, len(ds))
	for i, d := range ds {
		m, err := MessageFromDTO(d)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		res[i] = m
	}
	return res, nil
}

func messageToDTO(m bots.Message) dto.Message {
	return dto.Message{
		Text: m.Text(),
	}
}

func messagesToDTO(msgs []bots.Message) []dto.Message {
	res := make([]dto.Message, len(msgs))
	for i, m := range msgs {
		res[i] = messageToDTO(m)
	}
	return res
}
