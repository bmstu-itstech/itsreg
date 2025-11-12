package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewParticipant(t *testing.T) {
	// Вспомогательные функции для создания тестовых данных
	validID := func() bots.ParticipantID { return bots.NewParticipantID(1, "bot") }
	zeroID := func() bots.ParticipantID { return bots.ParticipantID{} }

	tests := []struct {
		name        string
		id          bots.ParticipantID
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "Valid participant",
			id:      validID(),
			wantErr: false,
		},
		{
			name:        "Zero id",
			id:          zeroID(),
			wantErr:     true,
			expectedErr: "id is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bots.NewParticipant(tt.id)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.id, got.ID())
				thr := got.ActiveThread()
				require.Nil(t, thr)
			}
		})
	}
}

func TestParticipant_StartThread(t *testing.T) {
	id := bots.NewParticipantID(1, "bot")
	prt := bots.MustNewParticipant(id)

	startState := bots.MustNewState(1)
	entry := bots.MustNewEntry("start", startState)

	started, err := prt.StartThread(entry)
	require.NoError(t, err)
	require.NotNil(t, started)
	require.Equal(t, started.State(), entry.Start())

	current := prt.ActiveThread()
	require.NotNil(t, current)
	require.Equal(t, started, current)
}
