package mappers

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

func entryFromDTO(d dto.Entry) (bots.Entry, error) {
	var vErr shared.ValidationError

	start, err := bots.NewState(d.Start)
	if err != nil {
		vErr = vErr.AppendPrefixed(err, "start")
	}

	if err = vErr.OrError(); err != nil {
		return bots.Entry{}, err
	}

	return bots.NewEntry(bots.EntryKey(d.Key), start)
}

func EntriesFromDTOPrefixed(ds []dto.Entry, prefix string) ([]bots.Entry, error) {
	var vErr shared.ValidationError

	entries := make([]bots.Entry, len(ds))
	for i, d := range ds {
		entry, err := entryFromDTO(d)
		if err != nil {
			vErr = vErr.AppendPrefixed(err, fmt.Sprintf("%s[%d]", prefix, i))
		}
		entries[i] = entry
	}

	if vErr.OrError() != nil {
		return nil, vErr
	}
	return entries, nil
}

func entryToDTO(e bots.Entry) dto.Entry {
	return dto.Entry{
		Key:   string(e.Key()),
		Start: e.Start().Int(),
	}
}

func entriesToDTO(es []bots.Entry) []dto.Entry {
	res := make([]dto.Entry, 0, len(es))
	for _, e := range es {
		res = append(res, entryToDTO(e))
	}
	return res
}
