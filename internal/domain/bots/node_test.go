package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewNode(t *testing.T) {
	edge := bots.NewEdge(bots.AlwaysTruePredicate{}, bots.State(2), bots.NoOp{})

	tests := []struct {
		name        string
		state       bots.State
		title       string
		edges       []bots.Edge
		msgs        []bots.Message
		opts        []bots.Option
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "Valid node with one message",
			state:   bots.State(1),
			title:   "test",
			edges:   []bots.Edge{edge},
			msgs:    []bots.Message{bots.MustNewMessage("test")},
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "Valid node with multiple messages and edges",
			state:   bots.State(1),
			title:   "test",
			edges:   []bots.Edge{edge, edge},
			msgs:    []bots.Message{bots.MustNewMessage("1"), bots.MustNewMessage("2")},
			opts:    nil,
			wantErr: false,
		},
		{
			name:        "Empty title - error",
			state:       bots.State(1),
			title:       "",
			edges:       []bots.Edge{edge},
			msgs:        []bots.Message{bots.MustNewMessage("test")},
			opts:        nil,
			wantErr:     true,
			expectedErr: "expected not empty title",
		},
		{
			name:        "Empty messages - error",
			state:       bots.State(1),
			title:       "test",
			edges:       []bots.Edge{},
			msgs:        []bots.Message{},
			wantErr:     true,
			expectedErr: "expected at least one message in node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := bots.NewNode(tt.state, tt.title, tt.edges, tt.msgs, tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				require.ErrorAs(t, err, &bots.InvalidInputError{})
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.state, node.State())
				require.Equal(t, tt.title, node.Title())
				require.Equal(t, tt.edges, node.Edges())
				require.Equal(t, tt.msgs, node.Messages())
			}
		})
	}
}

func TestNode_IsZero(t *testing.T) {
	msg := bots.MustNewMessage("some text")
	initialized := bots.MustNewNode(1, "test", nil, []bots.Message{msg}, nil)
	require.False(t, initialized.IsZero())

	var uninitialized bots.Node
	require.True(t, uninitialized.IsZero())
}

func TestNode_Transition(t *testing.T) {
	msg := bots.MustNewMessage("some text")

	t.Run("One edge - match", func(t *testing.T) {
		edge := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(2), bots.NoOp{})
		node := bots.MustNewNode(bots.State(1), "test", []bots.Edge{edge}, []bots.Message{msg}, nil)
		walked, ok := node.Transition(bots.MustNewMessage("a"))
		require.True(t, ok)
		require.Equal(t, edge, walked)
	})

	t.Run("One edge - no match", func(t *testing.T) {
		edge := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(2), bots.NoOp{})
		node := bots.MustNewNode(bots.State(1), "test", []bots.Edge{edge}, []bots.Message{msg}, nil)
		_, ok := node.Transition(bots.MustNewMessage("b"))
		require.False(t, ok)
	})

	t.Run("Two edges - unique match", func(t *testing.T) {
		edgeA := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(2), bots.NoOp{})
		edgeB := bots.NewEdge(bots.MustNewExactMatchPredicate("b"), bots.State(3), bots.NoOp{})
		node := bots.MustNewNode(bots.State(1), "test", []bots.Edge{edgeA, edgeB}, []bots.Message{msg}, nil)
		walked, ok := node.Transition(bots.MustNewMessage("b"))
		require.True(t, ok)
		require.Equal(t, edgeB, walked)
	})

	t.Run("Two edges - high priority match", func(t *testing.T) {
		edgeA1 := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(3), bots.NoOp{})
		edgeA2 := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(2), bots.NoOp{})
		node := bots.MustNewNode(bots.State(1), "test", []bots.Edge{edgeA1, edgeA2}, []bots.Message{msg}, nil)
		walked, ok := node.Transition(bots.MustNewMessage("a"))
		require.True(t, ok)
		require.Equal(t, edgeA1, walked)
	})

	t.Run("No edges - no match", func(t *testing.T) {
		node := bots.MustNewNode(bots.State(1), "test", nil, []bots.Message{msg}, nil)
		_, ok := node.Transition(bots.MustNewMessage("a"))
		require.False(t, ok)
	})
}

func TestNode_Children(t *testing.T) {
	msg := bots.MustNewMessage("some text")

	edge1 := bots.NewEdge(bots.MustNewExactMatchPredicate("b"), bots.State(3), bots.NoOp{})
	edge2 := bots.NewEdge(bots.MustNewExactMatchPredicate("a"), bots.State(2), bots.NoOp{})
	edge3 := bots.NewEdge(bots.MustNewExactMatchPredicate("b"), bots.State(3), bots.NoOp{})
	node := bots.MustNewNode(bots.State(1), "test", []bots.Edge{edge1, edge2, edge3}, []bots.Message{msg}, nil)

	// Упорядочивание происходит в порядке приоритета
	require.Equal(t, []bots.State{3, 2}, node.Children())
}
