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

type getRunsRepositoryStub struct {
	runs       []*bots.Run
	runsErr    error
	calls      int
	gotOwnerID bots.UserID
	gotFilter  port.RunsFilter
}

func (s *getRunsRepositoryStub) Run(context.Context, bots.RunID) (*bots.Run, error) {
	return nil, nil
}

func (s *getRunsRepositoryStub) RunsByOwnerID(
	_ context.Context, ownerID bots.UserID, filter port.RunsFilter,
) ([]*bots.Run, error) {
	s.calls++
	s.gotOwnerID = ownerID
	s.gotFilter = filter
	if s.runsErr != nil {
		return nil, s.runsErr
	}
	return s.runs, nil
}

func (s *getRunsRepositoryStub) ActiveRuns(context.Context) ([]*bots.Run, error) {
	return nil, nil
}

func (s *getRunsRepositoryStub) SaveRun(context.Context, *bots.Run) error {
	return nil
}

func (s *getRunsRepositoryStub) UpdateRun(context.Context, *bots.Run) error {
	return nil
}

func TestGetRunsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success maps runs and passes owner with both filters", func(t *testing.T) {
		status := "failed"
		botID := "b0001"
		repo := &getRunsRepositoryStub{
			runs: []*bots.Run{
				mustRestoreRun(t, "r0001", "b0001", bots.RunStatusFailed),
				mustRestoreRun(t, "r0002", "b0001", bots.RunStatusActive),
			},
		}
		h := query.NewGetRunsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetRunsRequest{ActorID: 10, Status: &status, BotID: &botID})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.RunStatusFailed, *repo.gotFilter.Status)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)

		require.Len(t, res, 2)
		require.Equal(t, "r0001", res[0].ID)
		require.Equal(t, "b0001", res[0].BotID)
		require.Equal(t, "failed", res[0].Status)
		require.Equal(t, "r0002", res[1].ID)
		require.Equal(t, "active", res[1].Status)
	})

	t.Run("without filters passes empty filter", func(t *testing.T) {
		repo := &getRunsRepositoryStub{runs: []*bots.Run{mustRestoreRun(t, "r0003", "b0002", bots.RunStatusStarting)}}
		h := query.NewGetRunsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetRunsRequest{ActorID: 77})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(77), repo.gotOwnerID)
		require.Nil(t, repo.gotFilter.Status)
		require.Nil(t, repo.gotFilter.BotID)
		require.Len(t, res, 1)
		require.Equal(t, "r0003", res[0].ID)
	})

	t.Run("invalid status returns empty response and skips repository", func(t *testing.T) {
		invalidStatus := "not-a-status"
		repo := &getRunsRepositoryStub{}
		h := query.NewGetRunsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetRunsRequest{ActorID: 10, Status: &invalidStatus})
		require.NoError(t, err)
		require.Empty(t, res)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		status := "active"
		repoErr := errors.New("db unavailable")
		repo := &getRunsRepositoryStub{runsErr: repoErr}
		h := query.NewGetRunsHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetRunsRequest{ActorID: 12, Status: &status})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(12), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.RunStatusActive, *repo.gotFilter.Status)
		require.Nil(t, repo.gotFilter.BotID)
	})
}

func mustRestoreRun(t *testing.T, id string, botID string, status bots.RunStatus) *bots.Run {
	t.Helper()

	startedAt := time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	run, err := bots.RestoreRun(
		bots.RunID(id),
		bots.BotID(botID),
		bots.Token("token_"+botID),
		status,
		nil,
		&startedAt,
		nil,
	)
	require.NoError(t, err)
	return run
}
