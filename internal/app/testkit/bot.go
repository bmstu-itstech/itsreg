package testkit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func MustValidBot(t *testing.T, id string, ownerID int64, scriptID string, deleted bool) *bots.Bot {
	t.Helper()

	now := time.Date(2026, time.April, 16, 12, 0, 0, 0, time.UTC)
	var deletedAt *time.Time
	if deleted {
		d := now.Add(time.Minute)
		deletedAt = &d
	}

	bot, err := bots.RestoreBot(
		bots.BotID(id),
		bots.UserID(ownerID),
		bots.ScriptID(scriptID),
		"token",
		"desc",
		now,
		now,
		deletedAt,
	)
	require.NoError(t, err)
	return bot
}
