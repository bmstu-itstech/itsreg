package bots_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

const timeEps = time.Millisecond

func TestNewThread(t *testing.T) {
	t.Run("Valid entry", func(t *testing.T) {
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		thread, err := bots.NewThread("bot-1", bots.UserID(2), entry)
		require.NoError(t, err)
		require.NotNil(t, thread)
		require.NotZero(t, thread.ID())
		require.Equal(t, entry.Key(), thread.Key())
		require.Equal(t, entry.Start(), thread.State())
		require.Empty(t, thread.Answers())
		require.Less(t, time.Since(thread.StartedAt()), timeEps)
	})

	t.Run("Empty botID", func(t *testing.T) {
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		_, err := bots.NewThread("", bots.UserID(2), entry)
		require.Error(t, err)
	})

	t.Run("Empty userID", func(t *testing.T) {
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		_, err := bots.NewThread("bot-1", 0, entry)
		require.Error(t, err)
	})

	t.Run("Empty entry", func(t *testing.T) {
		var entry bots.Entry
		_, err := bots.NewThread("bot-1", bots.UserID(2), entry)
		require.Error(t, err)
	})
}

func TestThread_SaveAnswer(t *testing.T) {
	state1 := bots.MustNewState(1)
	entry := bots.MustNewEntry("start", state1)
	thread := bots.MustNewThread("bot-1", bots.UserID(2), entry)

	msgA := bots.MustNewMessage("a")
	thread.SaveAnswer(msgA)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgA, thread.Answers()[state1])

	msgB := bots.MustNewMessage("b")
	thread.SaveAnswer(msgB)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgB, thread.Answers()[state1])

	state2 := bots.MustNewState(2)
	thread.StepTo(state2)

	msgC := bots.MustNewMessage("c")
	thread.SaveAnswer(msgC)
	require.Len(t, thread.Answers(), 2)
	require.Equal(t, msgB, thread.Answers()[state1])
	require.Equal(t, msgC, thread.Answers()[state2])
}

func TestThread_AppendAnswer(t *testing.T) {
	state1 := bots.MustNewState(1)
	entry := bots.MustNewEntry("start", state1)
	thread := bots.MustNewThread("bot-1", bots.UserID(2), entry)

	msgA := bots.MustNewMessage("a")
	thread.AppendAnswer(msgA)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgA, thread.Answers()[state1])

	msgB := bots.MustNewMessage("b")
	thread.AppendAnswer(msgB)
	require.Len(t, thread.Answers(), 1)

	composed := msgA.Merge(msgB)
	require.Equal(t, composed, thread.Answers()[state1])

	state2 := bots.MustNewState(2)
	thread.StepTo(state2)

	msgC := bots.MustNewMessage("c")
	thread.SaveAnswer(msgC)
	require.Len(t, thread.Answers(), 2)
	require.Equal(t, composed, thread.Answers()[state1])
	require.Equal(t, msgC, thread.Answers()[state2])
}
