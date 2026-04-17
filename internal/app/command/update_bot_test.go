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
	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

type updateBotRepositoryStub struct {
	bot            *bots.Bot
	botErr         error
	updateErr      error
	botCalls       int
	updateCalls    int
	updated        *bots.Bot
	requestedBotID bots.BotID
}

func (s *updateBotRepositoryStub) Bot(_ context.Context, id bots.BotID) (*bots.Bot, error) {
	s.botCalls++
	s.requestedBotID = id
	if s.botErr != nil {
		return nil, s.botErr
	}
	return s.bot, nil
}

func (s *updateBotRepositoryStub) BotsByOwnerID(context.Context, bots.UserID) ([]*bots.Bot, error) {
	return nil, nil
}

func (s *updateBotRepositoryStub) SaveBot(context.Context, *bots.Bot) error {
	return nil
}

func (s *updateBotRepositoryStub) UpdateBot(_ context.Context, bot *bots.Bot) error {
	s.updateCalls++
	s.updated = bot
	return s.updateErr
}

type updateBotScriptMetaProviderStub struct {
	meta              dto.ScriptMeta
	metaErr           error
	scriptMetaCalls   int
	requestedScriptID bots.ScriptID
}

func (s *updateBotScriptMetaProviderStub) ScriptMeta(_ context.Context, id bots.ScriptID) (dto.ScriptMeta, error) {
	s.scriptMetaCalls++
	s.requestedScriptID = id
	if s.metaErr != nil {
		return dto.ScriptMeta{}, s.metaErr
	}
	return s.meta, nil
}

func TestUpdateBotHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{meta: validUpdateBotScriptMeta("sc0002", 42, false)}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		res, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.NoError(t, err)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, bots.BotID("b0001"), repo.requestedBotID)
		require.Equal(t, 1, smp.scriptMetaCalls)
		require.Equal(t, bots.ScriptID("sc0002"), smp.requestedScriptID)
		require.Equal(t, 1, repo.updateCalls)
		require.NotNil(t, repo.updated)
		require.Equal(t, "sc0002", repo.updated.ScriptID().String())
		require.Equal(t, "new-token", repo.updated.Token().String())
		require.Equal(t, "new desc", repo.updated.Desc())
		require.Equal(t, "b0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "sc0002", res.ScriptID)
		require.Equal(t, "new desc", res.Desc)
	})

	t.Run("bot not found", func(t *testing.T) {
		repo := &updateBotRepositoryStub{botErr: port.ErrBotNotFound}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("bot repository error", func(t *testing.T) {
		dbErr := errors.New("db unavailable")
		repo := &updateBotRepositoryStub{botErr: dbErr}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, dbErr)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("deleted bot is hidden as not found", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0002", 42, "sc0001", true)}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("foreign bot is hidden as not found", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0003", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		req := validUpdateBotRequest()
		req.ActorID = 10

		_, err := h.Handle(t.Context(), req)
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script not found", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{metaErr: port.ErrScriptNotFound}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script is deleted", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{meta: validUpdateBotScriptMeta("sc0002", 42, true)}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("script not found for foreign script owner", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{meta: validUpdateBotScriptMeta("sc0002", 10, false)}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), validUpdateBotRequest())
		require.ErrorIs(t, err, port.ErrScriptNotFound)
		require.Equal(t, 1, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("invalid token returns validation error", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), command.UpdateBotRequest{
			ActorID: 42,
			BotID:   "b0001",
			Token:   strPtr(""),
		})
		var vErr shared.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 0, smp.scriptMetaCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("invalid script id returns validation error", func(t *testing.T) {
		repo := &updateBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		smp := &updateBotScriptMetaProviderStub{meta: validUpdateBotScriptMeta("", 42, false)}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), command.UpdateBotRequest{
			ActorID:  42,
			BotID:    "b0001",
			ScriptID: strPtr(""),
		})
		var vErr shared.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 1, smp.scriptMetaCalls)
		require.Equal(t, bots.ScriptID(""), smp.requestedScriptID)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("token already exists", func(t *testing.T) {
		repo := &updateBotRepositoryStub{
			bot:       testkit.MustValidBot(t, "b0001", 42, "sc0001", false),
			updateErr: port.ErrTokenAlreadyExists,
		}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), command.UpdateBotRequest{
			ActorID: 42,
			BotID:   "b0001",
			Desc:    strPtr("updated"),
		})
		require.ErrorIs(t, err, port.ErrTokenAlreadyExists)
		require.Equal(t, 1, repo.updateCalls)
	})

	t.Run("update repository error", func(t *testing.T) {
		dbErr := errors.New("db unavailable")
		repo := &updateBotRepositoryStub{
			bot:       testkit.MustValidBot(t, "b0001", 42, "sc0001", false),
			updateErr: dbErr,
		}
		smp := &updateBotScriptMetaProviderStub{}
		h := command.NewUpdateBotHandler(repo, smp, logger)

		_, err := h.Handle(t.Context(), command.UpdateBotRequest{
			ActorID: 42,
			BotID:   "b0001",
			Desc:    strPtr("updated"),
		})
		require.ErrorIs(t, err, dbErr)
		require.Equal(t, 1, repo.updateCalls)
	})
}

func validUpdateBotRequest() command.UpdateBotRequest {
	return command.UpdateBotRequest{
		ActorID:  42,
		BotID:    "b0001",
		ScriptID: strPtr("sc0002"),
		Token:    strPtr("new-token"),
		Desc:     strPtr("new desc"),
	}
}

func validUpdateBotScriptMeta(id string, ownerID int64, deleted bool) dto.ScriptMeta {
	return dto.ScriptMeta{
		ID:      id,
		OwnerID: ownerID,
		Deleted: deleted,
	}
}

func strPtr(s string) *string {
	return &s
}
