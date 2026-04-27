package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getMailingsRepositoryStub struct {
	mailings    []*bots.Mailing
	mailingsErr error
	calls       int
	gotOwnerID  bots.UserID
	gotFilter   port.MailingsFilter
}

func (s *getMailingsRepositoryStub) Mailing(context.Context, bots.MailingID) (*bots.Mailing, error) {
	return nil, nil
}

func (s *getMailingsRepositoryStub) MailingsByOwnerID(
	_ context.Context,
	ownerID bots.UserID,
	filter port.MailingsFilter,
) ([]*bots.Mailing, error) {
	s.calls++
	s.gotOwnerID = ownerID
	s.gotFilter = filter
	if s.mailingsErr != nil {
		return nil, s.mailingsErr
	}
	return s.mailings, nil
}

func (s *getMailingsRepositoryStub) SaveMailing(context.Context, *bots.Mailing) error {
	return nil
}

func (s *getMailingsRepositoryStub) UpdateMailing(context.Context, *bots.Mailing) error {
	return nil
}

func TestGetMailingsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success maps mailings and passes owner with both filters", func(t *testing.T) {
		status := "failed"
		botID := "b0001"
		repo := &getMailingsRepositoryStub{
			mailings: []*bots.Mailing{
				mustRestoreMailing(t, "m0001", "b0001", bots.MailingStatusFailed),
				mustRestoreMailing(t, "m0002", "b0001", bots.MailingStatusStarted),
			},
		}
		h := query.NewGetMailingsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetMailingsRequest{ActorID: 10, Status: &status, BotID: &botID})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.MailingStatusFailed, *repo.gotFilter.Status)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)

		require.Len(t, res, 2)
		require.Equal(t, "m0001", res[0].ID)
		require.Equal(t, "b0001", res[0].BotID)
		require.Equal(t, "failed", res[0].Status)
		require.Equal(t, "m0002", res[1].ID)
		require.Equal(t, "started", res[1].Status)
	})

	t.Run("without filters passes empty filter", func(t *testing.T) {
		repo := &getMailingsRepositoryStub{
			mailings: []*bots.Mailing{mustRestoreMailing(t, "m0003", "b0002", bots.MailingStatusScheduled)},
		}
		h := query.NewGetMailingsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetMailingsRequest{ActorID: 77})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(77), repo.gotOwnerID)
		require.Nil(t, repo.gotFilter.Status)
		require.Nil(t, repo.gotFilter.BotID)
		require.Len(t, res, 1)
		require.Equal(t, "m0003", res[0].ID)
	})

	t.Run("invalid status returns empty response and skips repository", func(t *testing.T) {
		invalidStatus := "not-a-status"
		repo := &getMailingsRepositoryStub{}
		h := query.NewGetMailingsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetMailingsRequest{ActorID: 10, Status: &invalidStatus})
		require.NoError(t, err)
		require.Empty(t, res)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		status := "started"
		repoErr := errors.New("db unavailable")
		repo := &getMailingsRepositoryStub{mailingsErr: repoErr}
		h := query.NewGetMailingsHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetMailingsRequest{ActorID: 12, Status: &status})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(12), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.MailingStatusStarted, *repo.gotFilter.Status)
		require.Nil(t, repo.gotFilter.BotID)
	})
}

func mustRestoreMailing(t *testing.T, id string, botID string, status bots.MailingStatus) *bots.Mailing {
	t.Helper()

	createdAt := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	mailing, err := bots.RestoreMailing(
		bots.MailingID(id),
		bots.BotID(botID),
		"Mailing",
		bots.EntryKey("entry-1"),
		status,
		[]bots.UserID{101, 202},
		nil,
		createdAt,
		nil,
		nil,
	)
	require.NoError(t, err)
	return mailing
}
