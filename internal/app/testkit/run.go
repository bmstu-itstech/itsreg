package testkit

import (
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/stretchr/testify/require"
)

func MustNewRun(t *testing.T, id string, botID string, status bots.RunStatus) *bots.Run {
	t.Helper()

	now := time.Date(2026, time.September, 16, 30, 0, 0, 0, time.UTC)

	var startedAt *time.Time
	if status == bots.RunStatusActive || status == bots.RunStatusStopping || status == bots.RunStatusStopped {
		startedAt = &now
	}

	var stoppedAt *time.Time
	if status == bots.RunStatusStopped || status == bots.RunStatusFailed {
		stoppedAt = &now
	}

	run, err := bots.RestoreRun(
		bots.RunID(id),
		bots.BotID(botID),
		"token",
		status,
		nil,
		startedAt,
		stoppedAt,
	)
	require.NoError(t, err)
	return run
}
