package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getScriptRepositoryStub struct {
	script    *bots.Script
	scriptErr error
}

func (s *getScriptRepositoryStub) Script(_ context.Context, _ bots.ScriptID) (*bots.Script, error) {
	if s.scriptErr != nil {
		return nil, s.scriptErr
	}
	return s.script, nil
}

func (s *getScriptRepositoryStub) ScriptsByOwnerID(context.Context, bots.UserID) ([]*bots.Script, error) {
	return nil, nil
}

func (s *getScriptRepositoryStub) SaveScript(context.Context, *bots.Script) error {
	return nil
}

func (s *getScriptRepositoryStub) UpdateScript(context.Context, *bots.Script) error {
	return nil
}

func TestGetScriptHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		repo := &getScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, false)}
		h := query.NewGetScriptHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetScriptRequest{ActorID: 42, ScriptID: "sc0001"})
		require.NoError(t, err)
		require.Equal(t, "sc0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "desc", res.Desc)
	})

	t.Run("deleted script is hidden as not found", func(t *testing.T) {
		repo := &getScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0002", 42, true)}
		h := query.NewGetScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetScriptRequest{ActorID: 42, ScriptID: "sc0002"})
		require.ErrorIs(t, err, port.ErrScriptNotFound)
	})

	t.Run("foreign script is hidden as not found", func(t *testing.T) {
		repo := &getScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0003", 1, false)}
		h := query.NewGetScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetScriptRequest{ActorID: 42, ScriptID: "sc0003"})
		require.ErrorIs(t, err, port.ErrScriptNotFound)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		repo := &getScriptRepositoryStub{scriptErr: repoErr}
		h := query.NewGetScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetScriptRequest{ActorID: 42, ScriptID: "sc0004"})
		require.ErrorIs(t, err, repoErr)
	})
}
