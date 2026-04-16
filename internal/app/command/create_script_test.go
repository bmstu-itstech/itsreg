package command_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type createScriptRepositoryStub struct {
	saved     *bots.Script
	saveErr   error
	saveCalls int
}

func (s *createScriptRepositoryStub) Script(context.Context, bots.ScriptID) (*bots.Script, error) {
	return nil, nil
}

func (s *createScriptRepositoryStub) ScriptsByOwnerID(context.Context, bots.UserID) ([]*bots.Script, error) {
	return nil, nil
}

func (s *createScriptRepositoryStub) SaveScript(_ context.Context, script *bots.Script) error {
	s.saveCalls++
	s.saved = script
	return s.saveErr
}

func (s *createScriptRepositoryStub) UpdateScript(context.Context, *bots.Script) error {
	return nil
}

func TestCreateScriptHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		repo := &createScriptRepositoryStub{}
		h := command.NewCreateScriptHandler(repo, logger)

		res, err := h.Handle(t.Context(), validCreateScriptRequest())
		require.NoError(t, err)
		require.NotEmpty(t, res.ScriptID)
		require.Equal(t, 1, repo.saveCalls)
		require.NotNil(t, repo.saved)
		require.Equal(t, res.ScriptID, repo.saved.ID().String())
		require.Equal(t, bots.UserID(10), repo.saved.OwnerID())
		require.Equal(t, "new script", repo.saved.Desc())
	})

	t.Run("validation error is returned before save", func(t *testing.T) {
		repo := &createScriptRepositoryStub{}
		h := command.NewCreateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.CreateScriptRequest{
			ActorID: 10,
			Desc:    "new script",
			Nodes: []dto.Node{{
				State:    0,
				Title:    "start",
				Messages: []dto.Message{{Text: "hello"}},
			}},
			Entries: []dto.Entry{{Key: "start", Start: 1}},
		})
		require.Error(t, err)
		var vErr bots.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 0, repo.saveCalls)
	})

	t.Run("duplicate script is returned as-is", func(t *testing.T) {
		repo := &createScriptRepositoryStub{saveErr: port.ErrScriptAlreadyExists}
		h := command.NewCreateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validCreateScriptRequest())
		require.ErrorIs(t, err, port.ErrScriptAlreadyExists)
		require.Equal(t, 1, repo.saveCalls)
	})

	t.Run("repository error is returned as-is", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		repo := &createScriptRepositoryStub{saveErr: repoErr}
		h := command.NewCreateScriptHandler(repo, logger)

		_, err := h.Handle(t.Context(), validCreateScriptRequest())
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, repo.saveCalls)
	})
}

func validCreateScriptRequest() command.CreateScriptRequest {
	return command.CreateScriptRequest{
		ActorID: 10,
		Desc:    "new script",
		Nodes: []dto.Node{{
			State:    1,
			Title:    "start",
			Messages: []dto.Message{{Text: "hello"}},
		}},
		Entries: []dto.Entry{{Key: "start", Start: 1}},
	}
}
