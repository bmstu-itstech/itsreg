package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/stretchr/testify/require"
)

type getRunProviderStub struct {
	run    dto.OwnedRun
	runErr error
}

func (s *getRunProviderStub) OwnedRun(_ context.Context, _ bots.RunID) (dto.OwnedRun, error) {
	if s.runErr != nil {
		return dto.OwnedRun{}, s.runErr
	}
	return s.run, nil
}

func TestGetRunHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		req := query.GetRunRequest{ActorID: 42, RunID: "r0001"}
		repo := &getRunProviderStub{run: validOwnedRun(42)}
		h := query.NewGetRunHandler(repo, logger)

		res, err := h.Handle(t.Context(), req)
		require.NoError(t, err)
		require.Equal(t, "r0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "b0001", res.BotID)
		require.Equal(t, "failed", res.Status)
		require.NotNil(t, res.ErrorMsg)
		require.Equal(t, "Some error occurred", *res.ErrorMsg)
		require.NotNil(t, res.StartedAt)
		require.NotNil(t, res.StoppedAt)
	})

	t.Run("run not found", func(t *testing.T) {
		req := query.GetRunRequest{ActorID: 42, RunID: "r0001"}
		repo := &getRunProviderStub{runErr: port.ErrRunNotFound}
		h := query.NewGetRunHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrRunNotFound)
	})

	t.Run("foreign run is hidden as not found", func(t *testing.T) {
		req := query.GetRunRequest{ActorID: 42, RunID: "r0001"}
		repo := &getRunProviderStub{run: validOwnedRun(10)}
		h := query.NewGetRunHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrRunNotFound)
	})

	t.Run("provider error is returned", func(t *testing.T) {
		req := query.GetRunRequest{ActorID: 42, RunID: "r0001"}
		repoErr := errors.New("db unavailable")
		repo := &getRunProviderStub{runErr: repoErr}
		h := query.NewGetRunHandler(repo, logger)

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, repoErr)
	})
}

func validOwnedRun(ownerID int64) dto.OwnedRun {
	errMsg := "Some error occurred"
	startedAt := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	stoppedAt := time.Date(2026, 4, 11, 10, 45, 0, 0, time.UTC)
	return dto.OwnedRun{
		ID:        "r0001",
		OwnerID:   ownerID,
		BotID:     "b0001",
		Status:    "failed",
		ErrorMsg:  &errMsg,
		StartedAt: &startedAt,
		StoppedAt: &stoppedAt,
	}
}
