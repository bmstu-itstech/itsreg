package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestMailingStatusFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      bots.MailingStatus
		wantErr   bool
		errSubstr string
	}{
		{name: "scheduled", input: "scheduled", want: bots.MailingStatusScheduled},
		{name: "started", input: "started", want: bots.MailingStatusStarted},
		{name: "completed", input: "completed", want: bots.MailingStatusCompleted},
		{name: "failed", input: "failed", want: bots.MailingStatusFailed},
		{name: "unknown", input: "unknown", wantErr: true, errSubstr: "unknown mailing status: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bots.MailingStatusFromString(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errSubstr)
				require.True(t, got.IsZero())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
