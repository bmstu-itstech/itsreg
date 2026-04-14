package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func optionsFromDTOPrefixed(ds []string, prefix string) ([]bots.Option, error) {
	var vErr bots.ValidationError

	options := make([]bots.Option, len(ds))
	for i, d := range ds {
		option, err := bots.NewOption(d)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		options[i] = option
	}

	if vErr.OrError() != nil {
		return nil, vErr
	}
	return options, nil
}

func optionsToDTO(d []bots.Option) []string {
	res := make([]string, len(d))
	for i, o := range d {
		res[i] = o.String()
	}
	return res
}
