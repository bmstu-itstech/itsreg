package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewOperationFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bots.Operation
		wantErr  bool
	}{
		{
			name:     "No operation",
			input:    "noop",
			expected: bots.NoOp{},
		},
		{
			name:     "Save operation",
			input:    "save",
			expected: bots.SaveOp{},
		},
		{
			name:     "Append operation",
			input:    "append",
			expected: bots.AppendOp{},
		},
		{
			name:    "Invalid operation",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := bots.NewOperationFromString(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &bots.InvalidInputError{})
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, op)
			}
		})
	}
}

func TestNoOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.MustNewOperationFromString("noop")

	in := bots.NewMessage("op")
	op.Apply(thread, in)
	require.Empty(t, thread.Answers())
}

func TestSaveOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.MustNewOperationFromString("save")

	in1 := bots.NewMessage("op")
	op.Apply(thread, in1)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, in1, thread.Answers()[state])

	in2 := bots.NewMessage("b")
	op.Apply(thread, in2)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, in2, thread.Answers()[state])
}

func TestSaveAppendOp_Act(t *testing.T) {
	state := bots.State(1)
	entry := bots.MustNewEntry("start", state)
	thread := bots.MustNewThread(entry)
	op := bots.MustNewOperationFromString("append")

	in1 := bots.NewMessage("op")
	in2 := bots.NewMessage("b")
	op.Apply(thread, in1)
	op.Apply(thread, in2)

	expected := in1.Merge(in2)
	require.Len(t, thread.Answers(), 1)
	require.Contains(t, thread.Answers(), state)
	require.Equal(t, expected, thread.Answers()[state])
}
