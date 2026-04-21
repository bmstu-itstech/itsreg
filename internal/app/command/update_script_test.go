package command_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

type updateScriptRepositoryStub struct {
	script            *bots.Script
	scriptErr         error
	updateErr         error
	scriptCalls       int
	updateCalls       int
	updated           *bots.Script
	requestedScriptID bots.ScriptID
}

func (s *updateScriptRepositoryStub) Script(_ context.Context, id bots.ScriptID) (*bots.Script, error) {
	s.scriptCalls++
	s.requestedScriptID = id
	if s.scriptErr != nil {
		return nil, s.scriptErr
	}
	return s.script, nil
}

func (s *updateScriptRepositoryStub) ScriptsByOwnerID(context.Context, bots.UserID) ([]*bots.Script, error) {
	return nil, nil
}

func (s *updateScriptRepositoryStub) SaveScript(context.Context, *bots.Script) error {
	return nil
}

func (s *updateScriptRepositoryStub) UpdateScript(_ context.Context, script *bots.Script) error {
	s.updateCalls++
	s.updated = script
	return s.updateErr
}

func TestUpdateScriptHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, false)}
		h := command.NewUpdateScriptHandler(repo, logger)

		res, err := h.Handle(t.Context(), validUpdateScriptRequest())
		require.NoError(t, err)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, bots.ScriptID("sc0001"), repo.requestedScriptID)
		require.Equal(t, 1, repo.updateCalls)
		require.NotNil(t, repo.updated)
		require.Equal(t, "new script", repo.updated.Desc())
		require.Equal(t, "sc0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "new script", res.Desc)
		require.Len(t, res.Nodes, 1)
		require.Len(t, res.Entries, 1)
	})

	t.Run("dto mapping validation error is returned before repository call", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.UpdateScriptRequest{
			ActorID:  42,
			ScriptID: "sc0001",
			Desc:     "new script",
			Nodes: []dto.Node{{
				State:    0,
				Title:    "start",
				Messages: []dto.Message{{Text: "hello"}},
			}},
			Entries: []dto.Entry{{Key: "start", Start: 0}},
		})
		var vErr shared.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 0, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script not found", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{scriptErr: port.ErrScriptNotFound}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validUpdateScriptRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script repository error", func(t *testing.T) {
		dbErr := errors.New("db unavailable")
		repo := &updateScriptRepositoryStub{scriptErr: dbErr}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validUpdateScriptRequest())
		require.ErrorIs(t, err, dbErr)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("deleted script is hidden as not found", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, true)}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validUpdateScriptRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("foreign script is hidden as not found", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, false)}
		h := command.NewUpdateScriptHandler(repo, logger)

		req := validUpdateScriptRequest()
		req.ActorID = 10

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("invalid script returns validation error", func(t *testing.T) {
		repo := &updateScriptRepositoryStub{script: testkit.MustValidScript(t, "sc0001", 42, false)}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.UpdateScriptRequest{
			ActorID:  42,
			ScriptID: "sc0001",
			Desc:     "new script",
			Nodes: []dto.Node{{
				State:    2,
				Title:    "orphan",
				Messages: []dto.Message{{Text: "hello"}},
			}},
			Entries: []dto.Entry{{Key: "start", Start: 1}},
		})
		var vErr shared.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("update repository error", func(t *testing.T) {
		dbErr := errors.New("db unavailable")
		repo := &updateScriptRepositoryStub{
			script:    testkit.MustValidScript(t, "sc0001", 42, false),
			updateErr: dbErr,
		}
		h := command.NewUpdateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validUpdateScriptRequest())
		require.ErrorIs(t, err, dbErr)
		require.Equal(t, 1, repo.scriptCalls)
		require.Equal(t, 1, repo.updateCalls)
	})
}

func validUpdateScriptRequest() command.UpdateScriptRequest {
	return command.UpdateScriptRequest{
		ActorID:  42,
		ScriptID: "sc0001",
		Desc:     "new script",
		Nodes: []dto.Node{{
			State:    1,
			Title:    "start",
			Messages: []dto.Message{{Text: "hello"}},
		}},
		Entries: []dto.Entry{{Key: "start", Start: 1}},
	}
}
