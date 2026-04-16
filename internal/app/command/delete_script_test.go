package command_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type deleteScriptRepositoryStub struct {
	script      *bots.Script
	scriptErr   error
	updateErr   error
	scriptCalls int
	updateCalls int
	updated     *bots.Script
}

func (s *deleteScriptRepositoryStub) Script(_ context.Context, _ bots.ScriptID) (*bots.Script, error) {
	s.scriptCalls++
	if s.scriptErr != nil {
		return nil, s.scriptErr
	}
	return s.script, nil
}

func (s *deleteScriptRepositoryStub) ScriptsByOwnerID(context.Context, bots.UserID) ([]*bots.Script, error) {
	return nil, nil
}

func (s *deleteScriptRepositoryStub) SaveScript(context.Context, *bots.Script) error {
	return nil
}

func (s *deleteScriptRepositoryStub) UpdateScript(_ context.Context, script *bots.Script) error {
	s.updateCalls++
	s.updated = script
	return s.updateErr
}

func TestDeleteScriptHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("script not found", func(t *testing.T) {
		repo := &deleteScriptRepositoryStub{scriptErr: port.ErrScriptNotFound}
		h := command.NewDeleteScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteScriptRequest{ActorID: 42, ScriptID: "sc0001"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script already deleted", func(t *testing.T) {
		repo := &deleteScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, true)}
		h := command.NewDeleteScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteScriptRequest{ActorID: 42, ScriptID: "sc0001"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("foreign script is hidden as not found", func(t *testing.T) {
		repo := &deleteScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0002", 42, false)}
		h := command.NewDeleteScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteScriptRequest{ActorID: 1, ScriptID: "sc0002"})
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("active own script is deleted and updated", func(t *testing.T) {
		repo := &deleteScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0003", 42, false)}
		h := command.NewDeleteScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteScriptRequest{ActorID: 42, ScriptID: "sc0003"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.updateCalls)
		require.NotNil(t, repo.updated)
		require.True(t, repo.updated.Deleted())
	})
}
