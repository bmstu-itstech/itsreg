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

type getScriptsRepositoryStub struct {
	scripts    []*bots.Script
	scriptsErr error
	calls      int
	gotOwnerID bots.UserID
}

func (s *getScriptsRepositoryStub) Script(context.Context, bots.ScriptID) (*bots.Script, error) {
	return nil, nil
}

func (s *getScriptsRepositoryStub) ScriptsByOwnerID(_ context.Context, ownerID bots.UserID) ([]*bots.Script, error) {
	s.calls++
	s.gotOwnerID = ownerID
	if s.scriptsErr != nil {
		return nil, s.scriptsErr
	}
	return s.scripts, nil
}

func (s *getScriptsRepositoryStub) SaveScript(context.Context, *bots.Script) error {
	return nil
}

func (s *getScriptsRepositoryStub) UpdateScript(context.Context, *bots.Script) error {
	return nil
}

func TestGetScriptsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success returns mapped scripts and passes actor as owner", func(t *testing.T) {
		repo := &getScriptsRepositoryStub{
			scripts: []*bots.Script{
				testkit.MustValidScript(t, "sc0001", 10, false),
				testkit.MustValidScript(t, "sc0002", 10, false),
			},
		}
		h := query.NewGetScriptsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetScriptsRequest{ActorID: 10})
		require.NoError(t, err)

		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)

		require.Len(t, res, 2)
		require.Equal(t, "sc0001", res[0].ID)
		require.Equal(t, "sc0002", res[1].ID)
		require.Equal(t, int64(10), res[0].OwnerID)
		require.Equal(t, int64(10), res[1].OwnerID)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		repo := &getScriptsRepositoryStub{scriptsErr: repoErr}
		h := query.NewGetScriptsHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetScriptsRequest{ActorID: 10})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
	})

	t.Run("nil scripts from repository returns empty response", func(t *testing.T) {
		repo := &getScriptsRepositoryStub{scripts: nil}
		h := query.NewGetScriptsHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetScriptsRequest{ActorID: 10})
		require.NoError(t, err)
		require.Empty(t, res)
	})
}
