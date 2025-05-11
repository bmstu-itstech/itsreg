package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewBotMessage(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		opts        []bots.Option
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "Valid message without options",
			text:    "hello",
			opts:    []bots.Option{},
			wantErr: false,
		},
		{
			name:    "Valid message with options",
			text:    "hello",
			opts:    []bots.Option{"a", "b"},
			wantErr: false,
		},
		{
			name:    "Empty text message",
			text:    "",
			opts:    []bots.Option{},
			wantErr: false,
		},
		{
			name:        "Single empty option",
			text:        "hello",
			opts:        []bots.Option{""},
			wantErr:     true,
			expectedErr: "expected non-empty string options",
		},
		{
			name:        "Multiple options with one empty",
			text:        "hello",
			opts:        []bots.Option{"opt1", "", "opt2"},
			wantErr:     true,
			expectedErr: "expected non-empty string options",
		},
		{
			name:    "Nil options slice",
			text:    "hello",
			opts:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bots.NewBotMessage(tt.text, tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.text, got.Text())
				require.Equal(t, tt.opts, got.Options())
			}
		})
	}
}
