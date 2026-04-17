package command_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type deleteBotRepositoryStub struct {
	bot         *bots.Bot
	botErr      error
	updateErr   error
	botCalls    int
	updateCalls int
	updated     *bots.Bot
}

func (s *deleteBotRepositoryStub) Bot(_ context.Context, _ bots.BotID) (*bots.Bot, error) {
	s.botCalls++
	if s.botErr != nil {
		return nil, s.botErr
	}
	return s.bot, nil
}

func (s *deleteBotRepositoryStub) BotsByOwnerID(context.Context, bots.UserID) ([]*bots.Bot, error) {
	return nil, nil
}

func (s *deleteBotRepositoryStub) SaveBot(context.Context, *bots.Bot) error {
	return nil
}

func (s *deleteBotRepositoryStub) UpdateBot(_ context.Context, bot *bots.Bot) error {
	s.updateCalls++
	s.updated = bot
	return s.updateErr
}

func TestDeleteBotHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("bot not found", func(t *testing.T) {
		repo := &deleteBotRepositoryStub{botErr: port.ErrBotNotFound}
		h := command.NewDeleteBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteBotRequest{ActorID: 42, BotID: "b0001"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("bot already deleted", func(t *testing.T) {
		repo := &deleteBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", true)}
		h := command.NewDeleteBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteBotRequest{ActorID: 42, BotID: "b0001"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("foreign bot is hidden as not found", func(t *testing.T) {
		repo := &deleteBotRepositoryStub{bot: testkit.MustValidBot(t, "b0002", 42, "sc0001", false)}
		h := command.NewDeleteBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteBotRequest{ActorID: 1, BotID: "b0002"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, repo.botCalls)
		require.Equal(t, 0, repo.updateCalls)
	})

	t.Run("active own bot is deleted and updated", func(t *testing.T) {
		repo := &deleteBotRepositoryStub{bot: testkit.MustValidBot(t, "b0003", 42, "sc0001", false)}
		h := command.NewDeleteBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), command.DeleteBotRequest{ActorID: 42, BotID: "b0003"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.updateCalls)
		require.NotNil(t, repo.updated)
		require.True(t, repo.updated.Deleted())
	})
}
