package testkit

import (
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/stretchr/testify/require"
)

func MustValidScript(t *testing.T, id string, ownerID int64, deleted bool) *bots.Script {
	t.Helper()

	now := time.Date(2026, time.April, 16, 12, 0, 0, 0, time.UTC)
	var deletedAt *time.Time
	if deleted {
		d := now.Add(time.Minute)
		deletedAt = &d
	}

	script, err := bots.RestoreScript(
		bots.ScriptID(id),
		bots.UserID(ownerID),
		"desc",
		nil,
		nil,
		now,
		now,
		deletedAt,
	)
	require.NoError(t, err)
	return script
}
