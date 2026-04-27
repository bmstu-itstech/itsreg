package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getMailingProviderStub struct {
	mailing    dto.OwnedMailing
	mailingErr error
}

func (s *getMailingProviderStub) OwnedMailing(_ context.Context, _ bots.MailingID) (dto.OwnedMailing, error) {
	if s.mailingErr != nil {
		return dto.OwnedMailing{}, s.mailingErr
	}
	return s.mailing, nil
}

func TestGetMailingHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		req := query.GetMailingRequest{ActorID: 42, MailingID: "m0001"}
		repo := &getMailingProviderStub{mailing: validOwnedMailing(42)}
		h := query.NewGetMailingHandler(repo, logger)

		res, err := h.Handle(t.Context(), req)
		require.NoError(t, err)
		require.Equal(t, "m0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "b0001", res.BotID)
		require.Equal(t, "entry-1", res.EntryKey)
		require.Equal(t, "started", res.Status)
		require.Equal(t, []int64{101, 202}, res.Recipients)
		require.Equal(t, 1, res.SuccessCount)
		require.Equal(t, 0, res.FailCount)
		require.Equal(t, 1, res.TotalCount)
		require.NotNil(t, res.StartedAt)
	})

	t.Run("mailing not found", func(t *testing.T) {
		req := query.GetMailingRequest{ActorID: 42, MailingID: "m0001"}
		repo := &getMailingProviderStub{mailingErr: port.ErrMailingNotFound}
		h := query.NewGetMailingHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrMailingNotFound)
	})

	t.Run("foreign mailing is hidden as not found", func(t *testing.T) {
		req := query.GetMailingRequest{ActorID: 42, MailingID: "m0001"}
		repo := &getMailingProviderStub{mailing: validOwnedMailing(10)}
		h := query.NewGetMailingHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrMailingNotFound)
	})

	t.Run("provider error is returned", func(t *testing.T) {
		req := query.GetMailingRequest{ActorID: 42, MailingID: "m0001"}
		repoErr := errors.New("db unavailable")
		repo := &getMailingProviderStub{mailingErr: repoErr}
		h := query.NewGetMailingHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, repoErr)
	})
}

func validOwnedMailing(ownerID int64) dto.OwnedMailing {
	startedAt := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	return dto.OwnedMailing{
		ID:           "m0001",
		OwnerID:      ownerID,
		BotID:        "b0001",
		EntryKey:     "entry-1",
		Status:       "started",
		Recipients:   []int64{101, 202},
		Results:      []dto.UserMailingResult{{UserID: 101, Success: true}},
		SuccessCount: 1,
		FailCount:    0,
		TotalCount:   1,
		CreatedAt:    createdAt,
		StartedAt:    &startedAt,
		CompletedAt:  nil,
	}
}
