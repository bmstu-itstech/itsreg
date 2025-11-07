package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewBot(t *testing.T) {
	msg := bots.MustNewMessage("some text")
	node := bots.MustNewNode(1, "test", nil, []bots.Message{msg}, nil)
	entry := bots.MustNewEntry("start", 1)
	validScript := bots.MustNewScript([]bots.Node{node}, []bots.Entry{entry})
	zeroScript := bots.Script{}

	tests := []struct {
		name        string
		id          bots.BotID
		token       bots.Token
		author      bots.UserID
		script      bots.Script
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid bot",
			id:      "bot",
			token:   "token",
			author:  1,
			script:  validScript,
			wantErr: false,
		},
		{
			name:        "Empty bot id",
			id:          "",
			token:       "token",
			author:      1,
			script:      validScript,
			wantErr:     true,
			errContains: "expected not empty bot id",
		},
		{
			name:        "Empty token",
			id:          "bot",
			token:       "",
			author:      1,
			script:      validScript,
			wantErr:     true,
			errContains: "expected not empty bot token",
		},
		{
			name:        "Zero author id",
			id:          "bot",
			token:       "token",
			author:      0,
			script:      validScript,
			wantErr:     true,
			errContains: "expected not empty bot author id",
		},
		{
			name:        "Zero script",
			id:          "bot",
			token:       "token",
			author:      1,
			script:      zeroScript,
			wantErr:     true,
			errContains: "empty script",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := bots.NewBot(tt.id, tt.token, tt.author, tt.script)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.id, bot.ID())
				require.Equal(t, tt.token, bot.Token())
				require.Equal(t, tt.script, bot.Script())
			}
		})
	}
}
