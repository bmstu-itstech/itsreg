package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getBotRunsRepositoryStub struct {
	runs       []*bots.Run
	runsErr    error
	calls      int
	gotOwnerID bots.UserID
	gotFilter  port.RunsFilter
}

func (s *getBotRunsRepositoryStub) Run(context.Context, bots.RunID) (*bots.Run, error) {
	return nil, nil
}

func (s *getBotRunsRepositoryStub) RunsByOwnerID(
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

func (s *getBotRunsRepositoryStub) ActiveRuns(context.Context) ([]*bots.Run, error) {
	return nil, nil
}

func (s *getBotRunsRepositoryStub) SaveRun(context.Context, *bots.Run) error {
	return nil
}

func (s *getBotRunsRepositoryStub) UpdateRun(context.Context, *bots.Run) error {
	return nil
}

type botMetaProviderStub struct {
	meta     dto.BotMeta
	metaErr  error
	calls    int
	gotBotID bots.BotID
}

func (s *botMetaProviderStub) BotMeta(_ context.Context, id bots.BotID) (dto.BotMeta, error) {
	s.calls++
	s.gotBotID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

func TestGetBotRunsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success maps runs and passes bot id and status filters", func(t *testing.T) {
		status := "active"
		repo := &getBotRunsRepositoryStub{
			runs: []*bots.Run{
				mustRestoreRun(t, "r0001", "b0001", bots.StatusActive),
				mustRestoreRun(t, "r0002", "b0001", bots.StatusFailed),
			},
		}
		bmp := &botMetaProviderStub{meta: dto.BotMeta{ID: "b0001"}}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		res, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 10, BotID: "b0001", Status: &status})
		require.NoError(t, err)

		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0001"), bmp.gotBotID)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.StatusActive, *repo.gotFilter.Status)

		require.Len(t, res, 2)
		require.Equal(t, "r0001", res[0].ID)
		require.Equal(t, "b0001", res[0].BotID)
		require.Equal(t, "active", res[0].Status)
	})

	t.Run("without status passes bot id only", func(t *testing.T) {
		repo := &getBotRunsRepositoryStub{runs: []*bots.Run{mustRestoreRun(t, "r0003", "b0003", bots.StatusStarting)}}
		bmp := &botMetaProviderStub{meta: dto.BotMeta{ID: "b0003"}}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		res, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 77, BotID: "b0003"})
		require.NoError(t, err)

		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0003"), bmp.gotBotID)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(77), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0003"), *repo.gotFilter.BotID)
		require.Nil(t, repo.gotFilter.Status)
		require.Len(t, res, 1)
		require.Equal(t, "r0003", res[0].ID)
	})

	t.Run("invalid status returns empty response and skips providers", func(t *testing.T) {
		invalidStatus := "not-a-status"
		repo := &getBotRunsRepositoryStub{}
		bmp := &botMetaProviderStub{}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		res, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 10, BotID: "b0001", Status: &invalidStatus})
		require.NoError(t, err)
		require.Empty(t, res)
		require.Equal(t, 0, bmp.calls)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("bot not found is returned and repository is not called", func(t *testing.T) {
		repo := &getBotRunsRepositoryStub{}
		bmp := &botMetaProviderStub{metaErr: port.ErrBotNotFound}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 10, BotID: "b0099"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0099"), bmp.gotBotID)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("bot provider error is returned", func(t *testing.T) {
		providerErr := errors.New("provider unavailable")
		repo := &getBotRunsRepositoryStub{}
		bmp := &botMetaProviderStub{metaErr: providerErr}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 10, BotID: "b0001"})
		require.ErrorIs(t, err, providerErr)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("runs repository error is returned", func(t *testing.T) {
		status := "failed"
		repoErr := errors.New("db unavailable")
		repo := &getBotRunsRepositoryStub{runsErr: repoErr}
		bmp := &botMetaProviderStub{meta: dto.BotMeta{ID: "b0001"}}
		h := query.NewGetBotRunsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotRunsRequest{ActorID: 11, BotID: "b0001", Status: &status})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(11), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.StatusFailed, *repo.gotFilter.Status)
	})
}
