package bots_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name    string
		key     bots.EntryKey
		start   bots.State
		wantErr bool
		errCode string
	}{
		{
			name:    "Valid entry",
			key:     "start",
			start:   bots.MustNewState(1),
			wantErr: false,
		},
		{
			name:    "Empty key",
			key:     "",
			start:   bots.MustNewState(1),
			wantErr: true,
			errCode: "entry-empty-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := bots.NewEntry(tt.key, tt.start)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &bots.InvalidInputError{})
				var ierr bots.InvalidInputError
				ok := errors.As(err, &ierr)
				require.True(t, ok)
				require.Equal(t, tt.errCode, ierr.Code)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.key, entry.Key())
				require.Equal(t, tt.start, entry.Start())
			}
		})
	}
}
