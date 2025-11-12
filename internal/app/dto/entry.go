package dto

import "github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"

type Entry struct {
	Key   string
	Start int
}

func entryFromDto(dto Entry) (bots.Entry, error) {
	start, err := bots.NewState(dto.Start)
	if err != nil {
		return bots.Entry{}, err
	}
	return bots.NewEntry(bots.EntryKey(dto.Key), start)
}

func batchEntriesFromDto(dto []Entry) ([]bots.Entry, error) {
	res := make([]bots.Entry, 0, len(dto))
	for _, entry := range dto {
		e, err := entryFromDto(entry)
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func entryToDto(entry bots.Entry) Entry {
	return Entry{
		Key:   string(entry.Key()),
		Start: entry.Start().Int(),
	}
}

func batchEntriesToDto(entry []bots.Entry) []Entry {
	res := make([]Entry, 0, len(entry))
	for _, entry := range entry {
		res = append(res, entryToDto(entry))
	}
	return res
}
