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
				require.Empty(t, got.Threads())
				_, ok := got.CurrentThread()
				require.False(t, ok)
			}
		})
	}
}

func TestParticipant_Clone(t *testing.T) {
	id := bots.NewParticipantID(1, "bot")
	prt := bots.MustNewParticipant(id)

	startState := bots.State(1)
	entry := bots.MustNewEntry("start", startState)

	_, err := prt.StartThread(entry)
	require.NoError(t, err)

	cloned := prt.Clone()
	require.Equal(t, prt.ID(), cloned.ID())
	require.Equal(t, prt.Threads(), cloned.Threads())
	c1, ok1 := prt.CurrentThread()
	c2, ok2 := cloned.CurrentThread()
	require.Equal(t, c1, c2)
	require.Equal(t, ok1, ok2)

	// Проверяем, что произошло глубокое копирование
	_, err = prt.StartThread(entry)
	require.NoError(t, err)
	require.NotEqual(t, prt.Threads(), cloned.Threads())
	c1, _ = prt.CurrentThread()
	c2, _ = cloned.CurrentThread()
	require.NotEqual(t, c1, c2)
}

func TestParticipant_StartThread(t *testing.T) {
	id := bots.NewParticipantID(1, "bot")
	prt := bots.MustNewParticipant(id)

	startState := bots.State(1)
	entry := bots.MustNewEntry("start", startState)

	started, err := prt.StartThread(entry)
	require.NoError(t, err)
	require.NotNil(t, started)

	current, ok := prt.CurrentThread()
	require.True(t, ok)
	require.Equal(t, started, current)
}
