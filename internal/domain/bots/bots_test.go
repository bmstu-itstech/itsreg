package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type rawDetail struct {
	Field string
	Code  bots.ErrorCode
}

func requireValidationErrorDetails(t *testing.T, err error, expected []rawDetail) {
	t.Helper()

	var vErr bots.ValidationError
	require.ErrorAs(t, err, &vErr)

	gotRaw := make([]rawDetail, len(expected))
	for i, d := range vErr.Details {
		gotRaw[i] = rawDetail{d.Field, d.Code}
	}
	require.ElementsMatch(t, expected, gotRaw)
}
