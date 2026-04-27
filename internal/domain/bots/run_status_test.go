package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestStatusFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      bots.RunStatus
		wantErr   bool
		errSubstr string
	}{
		{name: "starting", input: "starting", want: bots.RunStatusStarting},
		{name: "active", input: "active", want: bots.RunStatusActive},
		{name: "stopping", input: "stopping", want: bots.RunStatusStopping},
		{name: "stopped", input: "stopped", want: bots.RunStatusStopped},
		{name: "failed", input: "failed", want: bots.RunStatusFailed},
		{name: "unknown", input: "unknown", wantErr: true, errSubstr: "unknown status: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bots.RunStatusFromString(tt.input)
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
