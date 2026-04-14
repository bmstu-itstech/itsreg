package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name     string
		key      bots.EntryKey
		start    bots.State
		wantErr  bool
		errCheck func(t *testing.T, err error)
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
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"key", bots.ErrorCodeEntryEmptyKey},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := bots.NewEntry(tt.key, tt.start)
			if tt.wantErr {
				require.Error(t, err)
				tt.errCheck(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.key, entry.Key())
			require.Equal(t, tt.start, entry.Start())
		})
	}
}
