package bots_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

const timeEps = time.Millisecond

func TestNewThread(t *testing.T) {
	t.Run("Valid entry", func(t *testing.T) {
		entry := bots.MustNewEntry("start", 1)
		thread, err := bots.NewThread(entry)
		require.NoError(t, err)
		require.NotNil(t, thread)
		require.NotZero(t, thread.Id())
		require.Equal(t, entry.Key(), thread.Key())
		require.Equal(t, entry.Start(), thread.State())
		require.Empty(t, thread.Answers())
		require.Less(t, time.Now().Sub(thread.StartedAt()), timeEps)
	})

	t.Run("Empty entry", func(t *testing.T) {
		var entry bots.Entry
		_, err := bots.NewThread(entry)
		require.Error(t, err)
	})
}

func TestThread_SaveAnswer(t *testing.T) {
	state1 := bots.State(1)
	entry := bots.MustNewEntry("start", state1)
	thread := bots.MustNewThread(entry)

	msgA := bots.MustNewMessage("a")
	thread.SaveAnswer(msgA)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgA, thread.Answers()[state1])

	msgB := bots.MustNewMessage("b")
	thread.SaveAnswer(msgB)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgB, thread.Answers()[state1])

	state2 := bots.State(2)
	thread.StepTo(state2)

	msgC := bots.MustNewMessage("c")
	thread.SaveAnswer(msgC)
	require.Len(t, thread.Answers(), 2)
	require.Equal(t, msgB, thread.Answers()[state1])
	require.Equal(t, msgC, thread.Answers()[state2])
}

func TestThread_AppendAnswer(t *testing.T) {
	state1 := bots.State(1)
	entry := bots.MustNewEntry("start", state1)
	thread := bots.MustNewThread(entry)

	msgA := bots.MustNewMessage("a")
	thread.AppendAnswer(msgA)
	require.Len(t, thread.Answers(), 1)
	require.Equal(t, msgA, thread.Answers()[state1])

	msgB := bots.MustNewMessage("b")
	thread.AppendAnswer(msgB)
	require.Len(t, thread.Answers(), 1)

	composed := msgA.Merge(msgB)
	require.Equal(t, composed, thread.Answers()[state1])

	state2 := bots.State(2)
	thread.StepTo(state2)

	msgC := bots.MustNewMessage("c")
	thread.SaveAnswer(msgC)
	require.Len(t, thread.Answers(), 2)
	require.Equal(t, composed, thread.Answers()[state1])
	require.Equal(t, msgC, thread.Answers()[state2])
}
