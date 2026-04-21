package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getBotsRepositoryStub struct {
	bots       []*bots.Bot
	botsErr    error
	calls      int
	gotOwnerID bots.UserID
}

func (s *getBotsRepositoryStub) Bot(context.Context, bots.BotID) (*bots.Bot, error) {
	return nil, nil
}

func (s *getBotsRepositoryStub) BotsByOwnerID(_ context.Context, ownerID bots.UserID) ([]*bots.Bot, error) {
	s.calls++
	s.gotOwnerID = ownerID
	if s.botsErr != nil {
		return nil, s.botsErr
	}
	return s.bots, nil
}

func (s *getBotsRepositoryStub) SaveBot(context.Context, *bots.Bot) error {
	return nil
}

func (s *getBotsRepositoryStub) UpdateBot(context.Context, *bots.Bot) error {
	return nil
}

func TestGetBotsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success returns mapped bots and passes actor as owner", func(t *testing.T) {
		repo := &getBotsRepositoryStub{
			bots: []*bots.Bot{
				testkit.MustValidBot(t, "b0001", 10, "sc0001", false),
				testkit.MustValidBot(t, "b0002", 10, "sc0002", false),
			},
		}
		h := query.NewGetBotsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetBotsRequest{ActorID: 10})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)

		require.Len(t, res, 2)
		require.Equal(t, "b0001", res[0].ID)
		require.Equal(t, "b0002", res[1].ID)
		require.Equal(t, int64(10), res[0].OwnerID)
		require.Equal(t, int64(10), res[1].OwnerID)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		repo := &getBotsRepositoryStub{botsErr: repoErr}
		h := query.NewGetBotsHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetBotsRequest{ActorID: 10})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
	})

	t.Run("nil bots from repository returns empty response", func(t *testing.T) {
		repo := &getBotsRepositoryStub{bots: nil}
		h := query.NewGetBotsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetBotsRequest{ActorID: 10})
		require.NoError(t, err)
		require.Empty(t, res)
	})
}
