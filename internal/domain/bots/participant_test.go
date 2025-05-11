package bots_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewParticipant(t *testing.T) {
	// Вспомогательные функции для создания тестовых данных
	validID := func() bots.ParticipantId { return bots.NewParticipantId(1, "bot") }
	zeroID := func() bots.ParticipantId { return bots.ParticipantId{} }

	tests := []struct {
		name        string
		id          bots.ParticipantId
		username    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:     "Valid participant",
			id:       validID(),
			username: "username",
			wantErr:  false,
		},
		{
			name:        "Zero id",
			id:          zeroID(),
			username:    "username",
			wantErr:     true,
			expectedErr: "id is zero",
		},
		{
			name:        "Empty username",
			id:          validID(),
			username:    "",
			wantErr:     true,
			expectedErr: "username is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bots.NewParticipant(tt.id, tt.username)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.id, got.Id())
				require.Equal(t, tt.username, got.Username())
				require.Empty(t, got.Threads())
				_, ok := got.CurrentThread()
				require.False(t, ok)
			}
		})
	}
}

func TestParticipant_StartThread(t *testing.T) {
	id := bots.NewParticipantId(1, "bot")
	prt := bots.MustNewParticipant(id, "username")

	startState := bots.State(1)
	entry := bots.MustNewEntry("start", startState)

	started, err := prt.StartThread(entry)
	require.NoError(t, err)
	require.NotNil(t, started)

	current, ok := prt.CurrentThread()
	require.True(t, ok)
	require.Equal(t, started, current)
}
