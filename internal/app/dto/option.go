package dto

import (
	"errors"
	"strconv"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type optionBuilder struct {
	state int
}

func (b *optionBuilder) WithState(state int) *optionBuilder {
	b.state = state
	return b
}

func (b *optionBuilder) Build(dto string) (bots.Option, error) {
	o, err := bots.NewOption(dto)
	if err != nil {
		return bots.Option{}, b.enrichError(err)
	}
	return o, nil
}

func (b *optionBuilder) BuildAll(dtos []string) ([]bots.Option, error) {
	var errs bots.MultiError
	res := make([]bots.Option, len(dtos))
	for i, dto := range dtos {
		o, err := b.Build(dto)
		if err != nil {
			errs.ExtendOrAppend(err)
		}
		res[i] = o
	}
	if errs.HasError() {
		return res, &errs
	}
	return res, nil
}

func (b *optionBuilder) enrichError(err error) error {
	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		if b.state != 0 {
			iiErr.Details["state"] = strconv.Itoa(b.state)
		}
		return iiErr
	}
	return err
}

func batchOptionsToDTO(dto []bots.Option) []string {
	res := make([]string, len(dto))
	for i, o := range dto {
		res[i] = o.String()
	}
	return res
}
