package bots_test

import (
	"errors"
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestNewBot(t *testing.T) {
	tests := []struct {
		name     string
		ownerID  bots.UserID
		scriptID bots.ScriptID
		token    bots.Token
		desc     string
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name:     "valid bot",
			ownerID:  1,
			scriptID: "script-1",
			token:    "token-1",
			desc:     "bot description",
		},
		{
			name:     "zero owner id",
			ownerID:  0,
			scriptID: "script-1",
			token:    "token-1",
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "ownerID")
			},
		},
		{
			name:    "zero script id",
			ownerID: 1, scriptID: "",
			token:   "token-1",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"scriptID", bots.ErrorCodeBotEmptyScriptID},
				})
			},
		},
		{
			name:     "zero token",
			ownerID:  1,
			scriptID: "script-1",
			token:    "",
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"token", bots.ErrorCodeBotEmptyToken},
				})
			},
		},
		{
			name:     "both zero scriptID and token",
			ownerID:  1,
			scriptID: "",
			token:    "",
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"scriptID", bots.ErrorCodeBotEmptyScriptID},
					{"token", bots.ErrorCodeBotEmptyToken},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := bots.NewBot(tt.ownerID, tt.scriptID, tt.token, tt.desc)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, bot)
				tt.errCheck(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, bot)
			require.Equal(t, tt.ownerID, bot.OwnerID())
			require.Equal(t, tt.scriptID, bot.ScriptID())
			require.Equal(t, tt.token, bot.Token())
			require.Equal(t, tt.desc, bot.Desc())
			require.False(t, bot.Deleted())
			require.False(t, bot.CreatedAt().IsZero())
			require.False(t, bot.UpdatedAt().IsZero())
		})
	}
}

func TestRestoreBot(t *testing.T) {
	createdAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(5 * time.Minute)
	deletedAt := createdAt.Add(10 * time.Minute)
	zeroTime := time.Time{}

	tests := []struct {
		name      string
		id        bots.BotID
		ownerID   bots.UserID
		scriptID  bots.ScriptID
		token     bots.Token
		desc      string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
		wantErr   bool
		errText   string
	}{
		{
			name:      "valid active bot",
			id:        "bot-1",
			ownerID:   1,
			scriptID:  "script-1",
			token:     "token-1",
			desc:      "bot",
			createdAt: createdAt,
			updatedAt: updatedAt,
		},
		{
			name:      "valid deleted bot",
			id:        "bot-1",
			ownerID:   1,
			scriptID:  "script-1",
			token:     "token-1",
			desc:      "bot",
			createdAt: createdAt,
			updatedAt: updatedAt,
			deletedAt: &deletedAt,
		},
		{
			name: "zero id",
			// id
			ownerID:   1,
			scriptID:  "script-1",
			token:     "token-1",
			createdAt: createdAt,
			updatedAt: updatedAt,
			wantErr:   true,
			errText:   "id is zero",
		},
		{
			name: "zero owner id",
			id:   "bot-1",
			// ownerID
			scriptID:  "script-1",
			token:     "token-1",
			createdAt: createdAt,
			updatedAt: updatedAt,
			wantErr:   true,
			errText:   "ownerID is zero",
		},
		{
			name: "zero script id",
			id:   "bot-1",
			// scriptID
			ownerID:   1,
			token:     "token-1",
			createdAt: createdAt,
			updatedAt: updatedAt,
			wantErr:   true,
			errText:   "scriptID is zero",
		},
		{
			name:    "zero token",
			id:      "bot-1",
			ownerID: 1,
			// token
			scriptID:  "script-1",
			createdAt: createdAt,
			updatedAt: updatedAt,
			wantErr:   true,
			errText:   "token is zero",
		},
		{
			name:     "zero created at",
			id:       "bot-1",
			ownerID:  1,
			scriptID: "script-1",
			token:    "token-1",
			// createdAt
			updatedAt: updatedAt,
			wantErr:   true,
			errText:   "createdAt is zero",
		},
		{
			name:      "zero updated at",
			id:        "bot-1",
			ownerID:   1,
			scriptID:  "script-1",
			token:     "token-1",
			createdAt: createdAt,
			// updatedAt
			wantErr: true,
			errText: "updatedAt is zero",
		},
		{
			name:      "zero deleted at",
			id:        "bot-1",
			ownerID:   1,
			scriptID:  "script-1",
			token:     "token-1",
			createdAt: createdAt,
			updatedAt: updatedAt,
			deletedAt: &zeroTime,
			wantErr:   true,
			errText:   "deletedAt is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := bots.RestoreBot(
				tt.id, tt.ownerID, tt.scriptID, tt.token, tt.desc, tt.createdAt, tt.updatedAt, tt.deletedAt,
			)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.errText)
				require.Nil(t, bot)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, bot)
			require.Equal(t, tt.id, bot.ID())
			require.Equal(t, tt.ownerID, bot.OwnerID())
			require.Equal(t, tt.scriptID, bot.ScriptID())
			require.Equal(t, tt.token, bot.Token())
			require.Equal(t, tt.desc, bot.Desc())
			require.Equal(t, tt.createdAt, bot.CreatedAt())
			require.Equal(t, tt.updatedAt, bot.UpdatedAt())
			if tt.deletedAt == nil {
				require.Nil(t, bot.DeletedAt())
			} else {
				require.NotNil(t, bot.DeletedAt())
				require.Equal(t, *tt.deletedAt, *bot.DeletedAt())
			}
		})
	}
}

func TestBot_EnsureActive(t *testing.T) {
	deletedAt := time.Date(2026, time.April, 11, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name    string
		bot     *bots.Bot
		wantErr bool
	}{
		{
			name: "active bot",
			bot:  bots.MustNewBot(1, "script-1", "token-1", "bot"),
		},
		{
			name: "deleted bot",
			bot: func() *bots.Bot {
				b, err := bots.RestoreBot(
					"bot-1", 1, "script-1", "token-1", "bot", deletedAt.Add(-time.Hour), deletedAt, &deletedAt,
				)
				require.NoError(t, err)
				return b
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.EnsureActive()
			if tt.wantErr {
				require.ErrorIs(t, err, bots.ErrBotDeleted)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBot_EnsureOwnedBy(t *testing.T) {
	bot := bots.MustNewBot(1, "script-1", "token-1", "bot")
	tests := []struct {
		name    string
		userID  bots.UserID
		wantErr bool
	}{
		{
			name:   "owner matches",
			userID: 1,
		},
		{
			name:    "owner mismatch",
			userID:  2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.EnsureOwnedBy(tt.userID)
			if tt.wantErr {
				require.ErrorIs(t, err, shared.ErrPermissionDenied)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBot_Delete(t *testing.T) {
	tests := []struct {
		name    string
		bot     func() *bots.Bot
		wantErr bool
	}{
		{
			name: "delete active bot",
			bot:  func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
		},
		{
			name: "delete already deleted bot",
			bot: func() *bots.Bot {
				b := bots.MustNewBot(1, "script-1", "token-1", "bot")
				require.NoError(t, b.Delete())
				return b
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := tt.bot()
			deletedBefore := bot.Deleted()
			err := bot.Delete()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, bots.ErrBotDeleted)
				require.Equal(t, deletedBefore, bot.Deleted())
				return
			}

			require.NoError(t, err)
			require.True(t, bot.Deleted())
			require.NotNil(t, bot.DeletedAt())
		})
	}
}

func TestBot_SetScriptID(t *testing.T) {
	tests := []struct {
		name     string
		bot      func() *bots.Bot
		scriptID bots.ScriptID
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name:     "update active bot",
			bot:      func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
			scriptID: "script-2",
		},
		{
			name:     "zero script id",
			bot:      func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
			scriptID: "",
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"scriptID", bots.ErrorCodeBotEmptyScriptID},
				})
			},
		},
		{
			name: "deleted bot",
			bot: func() *bots.Bot {
				b := bots.MustNewBot(1, "script-1", "token-1", "bot")
				require.NoError(t, b.Delete())
				return b
			},
			scriptID: "script-2",
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				require.ErrorIs(t, err, bots.ErrBotDeleted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := tt.bot()
			err := bot.SetScriptID(tt.scriptID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.scriptID, bot.ScriptID())
		})
	}
}

func TestBot_SetToken(t *testing.T) {
	tests := []struct {
		name     string
		bot      func() *bots.Bot
		token    bots.Token
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name:  "update active bot",
			bot:   func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
			token: "token-2",
		},
		{
			name:    "zero token",
			bot:     func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
			token:   "",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				requireValidationErrorDetails(t, err, []rawDetail{
					{"token", bots.ErrorCodeBotEmptyToken},
				})
			},
		},
		{
			name: "deleted bot",
			bot: func() *bots.Bot {
				b := bots.MustNewBot(1, "script-1", "token-1", "bot")
				require.NoError(t, b.Delete())
				return b
			},
			token:   "token-2",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				require.ErrorIs(t, err, bots.ErrBotDeleted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := tt.bot()
			err := bot.SetToken(tt.token)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.token, bot.Token())
		})
	}
}

func TestBot_SetDesc(t *testing.T) {
	tests := []struct {
		name     string
		bot      func() *bots.Bot
		desc     string
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name: "update active bot",
			bot:  func() *bots.Bot { return bots.MustNewBot(1, "script-1", "token-1", "bot") },
			desc: "updated bot",
		},
		{
			name: "deleted bot",
			bot: func() *bots.Bot {
				b := bots.MustNewBot(1, "script-1", "token-1", "bot")
				require.NoError(t, b.Delete())
				return b
			},
			desc:    "updated bot",
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				require.ErrorIs(t, err, bots.ErrBotDeleted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := tt.bot()
			err := bot.SetDesc(tt.desc)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.desc, bot.Desc())
		})
	}
}

func TestBot_Deleted(t *testing.T) {
	active := bots.MustNewBot(1, "script-1", "token-1", "bot")
	require.False(t, active.Deleted())

	deleted := bots.MustNewBot(1, "script-1", "token-1", "bot")
	require.NoError(t, deleted.Delete())
	require.True(t, deleted.Deleted())
	require.NotNil(t, deleted.DeletedAt())
	require.True(t, errors.Is(deleted.Delete(), bots.ErrBotDeleted))
}
