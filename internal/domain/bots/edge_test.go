package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewRegexpEdge(t *testing.T) {
	to := bots.State(1)
	pr := bots.Priority(1)
	op := bots.NoOp{}

	tests := []struct {
		name        string
		regex       string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "Valid regex",
			regex:   `^hello`,
			wantErr: false,
		},
		{
			name:        "Invalid regex",
			regex:       `[a-z`, // Незакрытая скобка
			wantErr:     true,
			expectedErr: "failed to compile regexp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bots.NewRegexpEdge(tt.regex, to, pr, op)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &bots.InvalidInputError{})
				require.ErrorContains(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegexpEdge_Match(t *testing.T) {
	to := bots.State(1)
	pr := bots.Priority(1)
	op := bots.NoOp{}

	tests := []struct {
		name     string
		regexp   string
		message  bots.Message
		expected bool
	}{
		{
			name:     "Matching text",
			regexp:   `^hello`,
			message:  bots.NewMessage("hello"),
			expected: true,
		},
		{
			name:     "Non-matching text",
			regexp:   `^hello`,
			message:  bots.NewMessage("world hello"),
			expected: false,
		},
		{
			name:     "Empty text",
			regexp:   `^$`,
			message:  bots.NewMessage(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edge, err := bots.NewRegexpEdge(tt.regexp, to, pr, op)
			require.NoError(t, err)
			require.Equal(t, tt.expected, edge.Match(tt.message))
		})
	}
}
