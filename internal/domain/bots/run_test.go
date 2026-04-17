package bots_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestNewRun(t *testing.T) {
	tests := []struct {
		name    string
		botID   bots.BotID
		token   bots.Token
		wantErr bool
	}{
		{
			name:  "success",
			botID: bots.BotID("b0001"),
			token: bots.Token("token:b0001"),
		},
		{
			name:    "empty botID",
			botID:   bots.BotID(""),
			token:   bots.Token("token:"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run, err := bots.NewRun(tt.botID, tt.token)
			if tt.wantErr {
				require.Nil(t, run)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, run)
			require.Equal(t, tt.botID, run.BotID())
			require.Equal(t, tt.token, run.Token())
			require.Equal(t, bots.StatusStarting, run.Status())
			require.Nil(t, run.ErrorMsg())
			require.Nil(t, run.StartedAt())
			require.Nil(t, run.StoppedAt())
			events := run.PullEvents()
			require.Len(t, events, 1)
			require.Equal(t, "run.start_requested", events[0].EventName())
		})
	}
}

func TestRestoreRun(t *testing.T) {
	startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(5 * time.Minute)
	zeroTime := time.Time{}
	msg := "something went wrong"

	tests := []struct {
		name      string
		id        bots.RunID
		botID     bots.BotID
		token     bots.Token
		status    bots.Status
		errorMsg  *string
		startedAt *time.Time
		stoppedAt *time.Time
		wantErr   bool
		errText   string
	}{
		{
			name:      "valid active run",
			id:        bots.RunID("r00001"),
			botID:     bots.BotID("b0001"),
			token:     bots.Token("token:b0001"),
			status:    bots.StatusActive,
			startedAt: &startedAt,
		},
		{
			name:      "valid failed run",
			id:        bots.RunID("r00002"),
			botID:     bots.BotID("b0002"),
			token:     bots.Token("token:b0002"),
			status:    bots.StatusFailed,
			errorMsg:  &msg,
			startedAt: &startedAt,
			stoppedAt: &stoppedAt,
		},
		{
			name:   "valid starting run",
			id:     bots.RunID("r00003"),
			botID:  bots.BotID("b0003"),
			token:  bots.Token("token:b0003"),
			status: bots.StatusStarting,
		},
		{
			name:      "zero run id",
			botID:     bots.BotID("b0001"),
			token:     bots.Token("token:b0001"),
			status:    bots.StatusActive,
			startedAt: &startedAt,
			wantErr:   true,
			errText:   "zero run id",
		},
		{
			name:      "zero token",
			id:        bots.RunID("r00004"),
			botID:     bots.BotID("b0001"),
			status:    bots.StatusActive,
			startedAt: &startedAt,
			wantErr:   true,
			errText:   "zero token",
		},
		{
			name:      "zero bot id",
			id:        bots.RunID("r00005"),
			token:     bots.Token("token:b0001"),
			status:    bots.StatusActive,
			startedAt: &startedAt,
			wantErr:   true,
			errText:   "zero botID",
		},
		{
			name:      "zero status",
			id:        bots.RunID("r00006"),
			botID:     bots.BotID("b0001"),
			token:     bots.Token("token:b0001"),
			startedAt: &startedAt,
			wantErr:   true,
			errText:   "zero status",
		},
		{
			name:      "zero started at",
			id:        bots.RunID("r00007"),
			botID:     bots.BotID("b0001"),
			token:     bots.Token("token:b0001"),
			status:    bots.StatusActive,
			startedAt: &zeroTime,
			wantErr:   true,
			errText:   "zero startedAt",
		},
		{
			name:      "zero stopped at",
			id:        bots.RunID("r00008"),
			botID:     bots.BotID("b0001"),
			token:     bots.Token("token:b0001"),
			status:    bots.StatusStopped,
			startedAt: &startedAt,
			stoppedAt: &zeroTime,
			wantErr:   true,
			errText:   "zero stoppedAt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run, err := bots.RestoreRun(
				tt.id,
				tt.botID,
				tt.token,
				tt.status,
				tt.errorMsg,
				tt.startedAt,
				tt.stoppedAt,
			)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.errText)
				require.Nil(t, run)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, run)
			require.Equal(t, tt.id, run.ID())
			require.Equal(t, tt.botID, run.BotID())
			require.Equal(t, tt.token, run.Token())
			require.Equal(t, tt.status, run.Status())

			if tt.errorMsg == nil {
				require.Nil(t, run.ErrorMsg())
			} else {
				require.NotNil(t, run.ErrorMsg())
				require.Equal(t, *tt.errorMsg, *run.ErrorMsg())
			}

			if tt.startedAt == nil {
				require.Nil(t, run.StartedAt())
			} else {
				require.NotNil(t, run.StartedAt())
				require.Equal(t, *tt.startedAt, *run.StartedAt())
			}

			if tt.stoppedAt == nil {
				require.Nil(t, run.StoppedAt())
			} else {
				require.NotNil(t, run.StoppedAt())
				require.Equal(t, *tt.stoppedAt, *run.StoppedAt())
			}

			require.Empty(t, run.PullEvents())
		})
	}
}

func TestRun_Start(t *testing.T) {
	tests := []struct {
		name            string
		prepare         func(t *testing.T) *bots.Run
		wantErr         bool
		wantEventName   string
		wantStartedAt   bool
		wantStoppedAt   bool
		wantErrorMsgNil bool
	}{
		{
			name: "starting -> active",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b0001", "token:b0001")
				require.NoError(t, err)
				_ = run.PullEvents() // сброс run.start_requested
				return run
			},
			wantEventName:   "run.started",
			wantStartedAt:   true,
			wantStoppedAt:   false,
			wantErrorMsgNil: true,
		},
		{
			name: "active -> active",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b0002", "token:b0002")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents() // сброс события первого Start
				return run
			},
			wantErr: true,
		},
		{
			name: "failed -> active",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b0003", "token:b0003")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Fail("boom"))
				_ = run.PullEvents()
				return run
			},
			wantErr: true,
		},
		{
			name: "stopped -> active",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b0004", "token:b0004")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents()
				require.NoError(t, run.Stop())
				return run
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := tt.prepare(t)

			err := run.Start()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
				return
			}

			require.NoError(t, err)
			require.Equal(t, bots.StatusActive, run.Status())

			if tt.wantStartedAt {
				require.NotNil(t, run.StartedAt())
				require.False(t, run.StartedAt().IsZero())
			}
			if !tt.wantStoppedAt {
				require.Nil(t, run.StoppedAt())
			}
			if tt.wantErrorMsgNil {
				require.Nil(t, run.ErrorMsg())
			}

			events := run.PullEvents()
			require.Len(t, events, 1)
			require.Equal(t, tt.wantEventName, events[0].EventName())
		})
	}
}

func TestRun_Fail(t *testing.T) {
	msg := "something went wrong"

	tests := []struct {
		name          string
		prepare       func(t *testing.T) *bots.Run
		wantErr       bool
		wantEventName string
	}{
		{
			name: "starting -> failed",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b1001", "token:b1001")
				require.NoError(t, err)
				_ = run.PullEvents() // сброс run.start_requested
				return run
			},
			wantEventName: "run.failed",
		},
		{
			name: "active -> failed",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b1002", "token:b1002")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents() // сброс run.started
				return run
			},
			wantEventName: "run.failed",
		},
		{
			name: "failed -> failed",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b1003", "token:b1003")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Fail("first fail"))
				_ = run.PullEvents()
				return run
			},
			wantErr: true,
		},
		{
			name: "stopped -> failed",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b1004", "token:b1004")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents()
				require.NoError(t, run.Stop())
				return run
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := tt.prepare(t)

			err := run.Fail(msg)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
				return
			}

			require.NoError(t, err)
			require.Equal(t, bots.StatusFailed, run.Status())
			require.NotNil(t, run.ErrorMsg())
			require.Equal(t, msg, *run.ErrorMsg())
			require.NotNil(t, run.StoppedAt())
			require.False(t, run.StoppedAt().IsZero())

			events := run.PullEvents()
			require.Len(t, events, 1)
			require.Equal(t, tt.wantEventName, events[0].EventName())
		})
	}
}

func TestRun_Stop(t *testing.T) {
	tests := []struct {
		name          string
		prepare       func(t *testing.T) *bots.Run
		wantErr       bool
		wantStoppedAt bool
	}{
		{
			name: "active -> stopped",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b2001", "token:b2001")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents() // сброс run.started
				return run
			},
			wantStoppedAt: true,
		},
		{
			name: "starting -> stopped",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b2002", "token:b2002")
				require.NoError(t, err)
				_ = run.PullEvents()
				return run
			},
			wantErr: true,
		},
		{
			name: "failed -> stopped",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b2003", "token:b2003")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Fail("boom"))
				_ = run.PullEvents()
				return run
			},
			wantErr: true,
		},
		{
			name: "stopped -> stopped",
			prepare: func(t *testing.T) *bots.Run {
				run, err := bots.NewRun("b2004", "token:b2004")
				require.NoError(t, err)
				_ = run.PullEvents()
				require.NoError(t, run.Start())
				_ = run.PullEvents()
				require.NoError(t, run.Stop())
				return run
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := tt.prepare(t)

			err := run.Stop()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
				return
			}

			require.NoError(t, err)
			require.Equal(t, bots.StatusStopped, run.Status())

			if tt.wantStoppedAt {
				require.NotNil(t, run.StoppedAt())
				require.False(t, run.StoppedAt().IsZero())
			}

			// В текущей реализации Stop не публикует событие.
			require.Empty(t, run.PullEvents())
		})
	}
}
