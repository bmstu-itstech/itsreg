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
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/stretchr/testify/require"
)

type createBotRepositoryStub struct {
	saved     *bots.Bot
	saveErr   error
	saveCalls int
}

func (s *createBotRepositoryStub) Bot(context.Context, bots.BotID) (*bots.Bot, error) {
	return nil, nil
}

func (s *createBotRepositoryStub) BotsByOwnerID(context.Context, bots.UserID) ([]*bots.Bot, error) {
	return nil, nil
}

func (s *createBotRepositoryStub) UpdateBot(context.Context, *bots.Bot) error {
	return nil
}

func (s *createBotRepositoryStub) SaveBot(_ context.Context, bot *bots.Bot) error {
	s.saved = bot
	s.saveCalls++
	return s.saveErr
}

type createBotScriptMetaProviderStub struct {
	meta    dto.ScriptMeta
	metaErr error
}

func (s *createBotScriptMetaProviderStub) ScriptMeta(context.Context, bots.ScriptID) (dto.ScriptMeta, error) {
	if s.metaErr != nil {
		return dto.ScriptMeta{}, s.metaErr
	}
	return s.meta, nil
}

func TestCreateBotHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(42, false)}
		h := command.NewCreateBotHandler(br, smp, logger)

		res, err := h.Handle(t.Context(), validCreateBotRequest())
		require.NoError(t, err)
		require.NotEmpty(t, res)

		require.Equal(t, 1, br.saveCalls)
		require.NotNil(t, br.saved)
		require.Equal(t, res.BotID, br.saved.ID().String())
		require.Equal(t, bots.UserID(42), br.saved.OwnerID())
		require.Equal(t, "sc0001", br.saved.ScriptID().String())
		require.Equal(t, "token", br.saved.Token().String())
		require.Equal(t, "desc", br.saved.Desc())
	})

	t.Run("script not found", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		smp := &createBotScriptMetaProviderStub{metaErr: port.ErrScriptNotFound}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 0, br.saveCalls)
	})

	t.Run("script is deleted", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(42, true)}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 0, br.saveCalls)
	})

	t.Run("permission denied for foreign script owner", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(10, false)}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, bots.ErrPermissionDenied)
		require.Equal(t, 0, br.saveCalls)
	})

	t.Run("create bot request is invalid", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(42, false)}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), command.CreateBotRequest{
			ActorID:  42,
			ScriptID: "sc0001",
			Token:    "", // Empty token
			Desc:     "desc",
		})
		var vErr bots.ValidationError
		require.ErrorAs(t, err, &vErr)
	})

	t.Run("bot already exists", func(t *testing.T) {
		br := &createBotRepositoryStub{saveErr: port.ErrBotAlreadyExists}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(42, false)}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, port.ErrBotAlreadyExists)
	})

	t.Run("script meta provider error", func(t *testing.T) {
		br := &createBotRepositoryStub{}
		dbError := errors.New("db unavailable")
		smp := &createBotScriptMetaProviderStub{metaErr: dbError}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, dbError)
		require.Equal(t, 0, br.saveCalls)
	})

	t.Run("save repository error", func(t *testing.T) {
		dbError := errors.New("db unavailable")
		br := &createBotRepositoryStub{saveErr: dbError}
		smp := &createBotScriptMetaProviderStub{meta: validScriptMeta(42, false)}
		h := command.NewCreateBotHandler(br, smp, logger)

		_, err := h.Handle(t.Context(), validCreateBotRequest())
		require.ErrorIs(t, err, dbError)
	})
}

func validCreateBotRequest() command.CreateBotRequest {
	return command.CreateBotRequest{
		ActorID:  42,
		ScriptID: "sc0001",
		Token:    "token",
		Desc:     "desc",
	}
}

func validScriptMeta(ownerID int64, deleted bool) dto.ScriptMeta {
	return dto.ScriptMeta{
		ID:      "sc0001",
		OwnerID: ownerID,
		Deleted: deleted,
	}
}
