package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestAlwaysTruePredicate(t *testing.T) {
	t.Run("always returns true", func(t *testing.T) {
		p := bots.AlwaysTruePredicate{}
		msg := bots.MustNewMessage("any text")
		require.True(t, p.Match(msg))
	})
}

func TestNewExactMatchPredicate(t *testing.T) {
	t.Run("Valid edge", func(t *testing.T) {
		e, err := bots.NewExactMatchPredicate("any text")
		require.NoError(t, err)
		require.NotZero(t, e)
	})

	t.Run("Empty text", func(t *testing.T) {
		_, err := bots.NewExactMatchPredicate("")
		require.Error(t, err)
		requireValidationErrorDetails(
			t, err, []rawDetail{{"text", bots.ErrorCodeExactMatchPredicateEmptyText}},
		)
	})
}

func TestExactMatchPredicate_Match(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected bool
	}{
		{
			name:     "Exact match",
			pattern:  "hello",
			text:     "hello",
			expected: true,
		},
		{
			name:     "Contains, no match",
			pattern:  "hello",
			text:     "hello world",
			expected: false,
		},
		{
			name:     "Case sensitive mismatch",
			pattern:  "Hello",
			text:     "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := bots.MustNewExactMatchPredicate(tt.pattern)
			msg := bots.MustNewMessage(tt.text)
			require.Equal(t, tt.expected, p.Match(msg))
		})
	}
}

func TestNewRegexMatchPredicate(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		wantErr  bool
		errCheck func(t *testing.T, err error)
	}{
		{
			name:    "Valid regex pattern",
			pattern: `^[a-z]+$`,
			wantErr: false,
		},
		{
			name:    "Empty pattern",
			pattern: "",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{{"pattern", bots.ErrorCodeRegexPredicateEmptyPattern}})
			},
		},
		{
			name:    "Invalid pattern",
			pattern: "^[a-z",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"pattern", bots.ErrorCodeRegexPredicateInvalidPattern},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := bots.NewRegexMatchPredicate(tt.pattern)
			if tt.wantErr {
				require.Error(t, err)
				tt.errCheck(t, err)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, p)
		})
	}
}

func TestRegexMatchPredicate_Match(t *testing.T) {
	t.Run("Non empty regex match", func(t *testing.T) {
		p := bots.MustNewRegexMatchPredicate(`^[a-z]+$`)

		msg := bots.MustNewMessage("hello")
		require.True(t, p.Match(msg))

		msg = bots.MustNewMessage("Hello")
		require.False(t, p.Match(msg))
	})
}
