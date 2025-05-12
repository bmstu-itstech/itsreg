package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNoOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.NoOp{}

	in := bots.MustNewMessage("op")
	op.Apply(thread, in)
	require.Empty(t, thread.Answers())
}

func TestSaveOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.SaveOp{}

	in1 := bots.MustNewMessage("op")
	op.Apply(thread, in1)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, in1, thread.Answers()[state])

	in2 := bots.MustNewMessage("b")
	op.Apply(thread, in2)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, in2, thread.Answers()[state])
}

func TestSaveAppendOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.AppendOp{}

	in1 := bots.MustNewMessage("op")
	in2 := bots.MustNewMessage("b")
	op.Apply(thread, in1)
	op.Apply(thread, in2)

	expected := in1.Merge(in2)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, expected, thread.Answers()[state])
}
