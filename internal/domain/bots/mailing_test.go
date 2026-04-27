package bots_test

import (
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/bmstu-itstech/itsreg/pkg/hlpr"
	"github.com/stretchr/testify/require"
)

func TestNewMailing(t *testing.T) {
	tests := []struct {
		name      string
		botID     bots.BotID
		mName     string
		entryKey  bots.EntryKey
		usersIn   []bots.UserID
		wantErr   bool
		errText   string
		isContain bool
	}{
		{
			name:     "success",
			botID:    bots.BotID("b0001"),
			mName:    "Mailing 1",
			entryKey: bots.EntryKey("entry-1"),
			usersIn:  []bots.UserID{1, 2},
		},
		{
			name:     "zero bot id",
			botID:    bots.BotID(""),
			mName:    "Mailing 2",
			entryKey: bots.EntryKey("entry-1"),
			usersIn:  []bots.UserID{1},
			wantErr:  true,
			errText:  "bot ID is zero",
		},
		{
			name:     "empty name",
			botID:    bots.BotID("b0001"),
			mName:    "",
			entryKey: bots.EntryKey("entry-1"),
			usersIn:  []bots.UserID{1},
			wantErr:  true,
			errText:  "validation error: mailing name is required;",
		},
		{
			name:     "zero entry key",
			botID:    bots.BotID("b0001"),
			mName:    "Mailing 2",
			entryKey: bots.EntryKey(""),
			usersIn:  []bots.UserID{1},
			wantErr:  true,
			errText:  "entry key is zero",
		},
		{
			name:      "empty user list",
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 3",
			entryKey:  bots.EntryKey("entry-1"),
			usersIn:   nil,
			wantErr:   true,
			errText:   "mailing should contain at least one recipient",
			isContain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailing, err := bots.NewMailing(tt.botID, tt.mName, tt.entryKey, tt.usersIn)
			if tt.wantErr {
				require.Error(t, err)
				if tt.isContain {
					require.ErrorContains(t, err, tt.errText)
				} else {
					require.EqualError(t, err, tt.errText)
				}
				require.Nil(t, mailing)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, mailing)

			events := mailing.PullEvents()
			require.Len(t, events, 1)
			require.Equal(t, "mailing.scheduled", events[0].EventName())
			require.False(t, events[0].OccurredAt().IsZero())
			require.Empty(t, mailing.PullEvents())
		})
	}
}

func TestRestoreMailing(t *testing.T) {
	createdAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
	startedAt := createdAt.Add(10 * time.Minute)
	completedAt := startedAt.Add(15 * time.Minute)
	zeroTime := time.Time{}
	msg := "something went wrong"

	tests := []struct {
		name        string
		id          bots.MailingID
		botID       bots.BotID
		mName       string
		entryKey    bots.EntryKey
		status      bots.MailingStatus
		usersIn     []bots.UserID
		resultsIn   []bots.UserMailingResult
		createdAt   time.Time
		startedAt   *time.Time
		completedAt *time.Time
		wantErr     bool
		errText     string
		errContains bool
	}{
		{
			name:      "valid scheduled mailing",
			id:        bots.MailingID("m00001"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 1",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusScheduled,
			usersIn:   []bots.UserID{1, 2},
			createdAt: createdAt,
		},
		{
			name:      "valid started mailing",
			id:        bots.MailingID("m00002"),
			botID:     bots.BotID("b0002"),
			mName:     "Mailing 2",
			entryKey:  bots.EntryKey("entry-2"),
			status:    bots.MailingStatusStarted,
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			startedAt: &startedAt,
		},
		{
			name:        "valid completed mailing",
			id:          bots.MailingID("m00003"),
			botID:       bots.BotID("b0003"),
			mName:       "Mailing 3",
			entryKey:    bots.EntryKey("entry-3"),
			status:      bots.MailingStatusCompleted,
			usersIn:     []bots.UserID{1, 2},
			resultsIn:   []bots.UserMailingResult{bots.NewSuccessMailingResult(1), bots.NewErrorMailingResult(2, msg)},
			createdAt:   createdAt,
			startedAt:   &startedAt,
			completedAt: &completedAt,
		},
		{
			name:      "zero id",
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 0",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusScheduled,
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			wantErr:   true,
			errText:   "id is zero",
		},
		{
			name:      "zero bot id",
			id:        bots.MailingID("m00004"),
			mName:     "Mailing 4",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusScheduled,
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			wantErr:   true,
			errText:   "bot id is zero",
		},
		{
			name:      "zero entry key",
			id:        bots.MailingID("m00005"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 5",
			status:    bots.MailingStatusScheduled,
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			wantErr:   true,
			errText:   "entry key is zero",
		},
		{
			name:      "zero status",
			id:        bots.MailingID("m00006"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 6",
			entryKey:  bots.EntryKey("entry-1"),
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			wantErr:   true,
			errText:   "status is zero",
		},
		{
			name:      "recipients is empty",
			id:        bots.MailingID("m00007"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 7",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusScheduled,
			createdAt: createdAt,
			wantErr:   true,
			errText:   "recipients is empty",
		},
		{
			name:      "zero created at",
			id:        bots.MailingID("m00008"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 8",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusScheduled,
			usersIn:   []bots.UserID{1},
			createdAt: zeroTime,
			wantErr:   true,
			errText:   "createdAt is zero",
		},
		{
			name:      "zero started at",
			id:        bots.MailingID("m00009"),
			botID:     bots.BotID("b0001"),
			mName:     "Mailing 9",
			entryKey:  bots.EntryKey("entry-1"),
			status:    bots.MailingStatusStarted,
			usersIn:   []bots.UserID{1},
			createdAt: createdAt,
			startedAt: &zeroTime,
			wantErr:   true,
			errText:   "startedAt is zero",
		},
		{
			name:        "zero completed at",
			id:          bots.MailingID("m00010"),
			botID:       bots.BotID("b0001"),
			mName:       "Mailing 10",
			entryKey:    bots.EntryKey("entry-1"),
			status:      bots.MailingStatusCompleted,
			usersIn:     []bots.UserID{1},
			createdAt:   createdAt,
			completedAt: &zeroTime,
			wantErr:     true,
			errText:     "completedAt is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailing, err := bots.RestoreMailing(
				tt.id,
				tt.botID,
				tt.mName,
				tt.entryKey,
				tt.status,
				tt.usersIn,
				tt.resultsIn,
				tt.createdAt,
				tt.startedAt,
				tt.completedAt,
			)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.errText)
				require.Nil(t, mailing)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, mailing)
			require.Empty(t, mailing.PullEvents())
		})
	}
}

func TestMailing_Started(t *testing.T) {
	startedAt := time.Date(2026, time.April, 11, 13, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(5 * time.Minute)

	tests := []struct {
		name    string
		prepare func(t *testing.T) *bots.Mailing
		wantErr bool
		errIs   error
	}{
		{
			name: "scheduled -> started",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.NewMailing("b0001", "Mailing 1", "entry-1", []bots.UserID{1})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				return mailing
			},
		},
		{
			name: "started -> started",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m10001",
					"b0001",
					"Mailing 1",
					"entry-1",
					bots.MailingStatusStarted,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					&completedAt,
				)
				require.NoError(t, err)
				return mailing
			},
			wantErr: true,
			errIs:   shared.ErrIllegalStateTransition,
		},
		{
			name: "completed -> started",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m10002",
					"b0001",
					"Mailing 2",
					"entry-1",
					bots.MailingStatusCompleted,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					&completedAt,
				)
				require.NoError(t, err)
				return mailing
			},
			wantErr: true,
			errIs:   shared.ErrIllegalStateTransition,
		},
		{
			name: "failed -> started",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m10003",
					"b0001",
					"Mailing 3",
					"entry-1",
					bots.MailingStatusFailed,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					&completedAt,
				)
				require.NoError(t, err)
				return mailing
			},
			wantErr: true,
			errIs:   shared.ErrIllegalStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailing := tt.prepare(t)
			err := mailing.MarkStarted()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errIs)
				return
			}

			require.NoError(t, err)
			require.NoError(t, mailing.MessageSent(1))
			require.Empty(t, mailing.PullEvents())
		})
	}
}

func TestMailing_Failed(t *testing.T) {
	startedAt := time.Date(2026, time.April, 11, 13, 10, 0, 0, time.UTC)
	completedAt := startedAt.Add(5 * time.Minute)

	tests := []struct {
		name    string
		prepare func(t *testing.T) *bots.Mailing
		wantErr bool
		errIs   error
	}{
		{
			name: "scheduled -> failed",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.NewMailing("b2001", "Mailing 1", "entry-1", []bots.UserID{1})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				return mailing
			},
		},
		{
			name: "started -> failed",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m20002",
					"b2002",
					"Mailing 2",
					"entry-2",
					bots.MailingStatusStarted,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					nil,
				)
				require.NoError(t, err)
				return mailing
			},
		},
		{
			name: "failed -> failed",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m20003",
					"b2003",
					"Mailing 3",
					"entry-3",
					bots.MailingStatusFailed,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					nil,
				)
				require.NoError(t, err)
				return mailing
			},
		},
		{
			name: "completed -> failed",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m20004",
					"b2004",
					"Mailing 4",
					"entry-4",
					bots.MailingStatusCompleted,
					[]bots.UserID{1},
					nil,
					startedAt,
					&startedAt,
					&completedAt,
				)
				require.NoError(t, err)
				return mailing
			},
			wantErr: true,
			errIs:   shared.ErrIllegalStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailing := tt.prepare(t)
			err := mailing.MarkFailed()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errIs)
				return
			}

			require.NoError(t, err)
			require.ErrorIs(t, mailing.MessageSent(1), bots.ErrMailingIsNotStarted)
		})
	}
}

func TestMailing_MessageSent(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T) *bots.Mailing
		userID  bots.UserID
		wantErr bool
		errIs   error
	}{
		{
			name: "not started",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.NewMailing("b3001", "Mailing 1", "entry-1", []bots.UserID{1})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				return mailing
			},
			userID:  1,
			wantErr: true,
			errIs:   bots.ErrMailingIsNotStarted,
		},
		{
			name: "user not in mailing",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.NewMailing("b3002", "Mailing 2", "entry-2", []bots.UserID{1})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				require.NoError(t, mailing.MarkStarted())
				return mailing
			},
			userID:  2,
			wantErr: true,
			errIs:   bots.ErrUserNotInRecipients,
		},
		{
			name: "all users delivered",
			prepare: func(t *testing.T) *bots.Mailing {
				mailing, err := bots.NewMailing("b3003", "Mailing 3", "entry-3", []bots.UserID{1, 2})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				require.NoError(t, mailing.MarkStarted())
				return mailing
			},
			userID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailing := tt.prepare(t)
			err := mailing.MessageSent(tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errIs)
				return
			}

			require.NoError(t, err)
			if tt.name == "all users delivered" {
				require.NoError(t, mailing.MessageSent(2))
				require.ErrorIs(t, mailing.MessageSent(1), bots.ErrMailingIsNotStarted)
			}
		})
	}
}

func TestMailing_SentCounters(t *testing.T) {
	createdAt := time.Date(2026, time.April, 11, 14, 0, 0, 0, time.UTC)
	startedAt := createdAt.Add(2 * time.Minute)

	tests := []struct {
		name             string
		mailing          *bots.Mailing
		wantSuccessful   int
		wantFailed       int
		wantTotal        int
		withMessageSends bool
	}{
		{
			name: "restored results",
			mailing: func() *bots.Mailing {
				mailing, err := bots.RestoreMailing(
					"m40001",
					"b4001",
					"Mailing 1",
					"entry-1",
					bots.MailingStatusCompleted,
					[]bots.UserID{1, 2},
					[]bots.UserMailingResult{
						bots.NewSuccessMailingResult(1),
						bots.NewErrorMailingResult(2, "boom"),
					},
					createdAt,
					&startedAt,
					hlpr.Ptr(startedAt.Add(3*time.Minute)),
				)
				require.NoError(t, err)
				return mailing
			}(),
			wantSuccessful: 1,
			wantFailed:     1,
			wantTotal:      2,
		},
		{
			name: "after successful sends",
			mailing: func() *bots.Mailing {
				mailing, err := bots.NewMailing("b4002", "Mailing 2", "entry-2", []bots.UserID{1, 2})
				require.NoError(t, err)
				_ = mailing.PullEvents()
				require.NoError(t, mailing.MarkStarted())
				require.NoError(t, mailing.MessageSent(1))
				require.NoError(t, mailing.MessageSent(2))
				return mailing
			}(),
			wantSuccessful: 2,
			wantFailed:     0,
			wantTotal:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantSuccessful, tt.mailing.SuccessCount())
			require.Equal(t, tt.wantFailed, tt.mailing.FailCount())
			require.Equal(t, tt.wantTotal, tt.mailing.RecipientsTotal())
		})
	}
}
