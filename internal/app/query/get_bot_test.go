package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/app/testkit"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/stretchr/testify/require"
)

type getBotRepositoryStub struct {
	bot    *bots.Bot
	botErr error
}

func (s *getBotRepositoryStub) Bot(_ context.Context, _ bots.BotID) (*bots.Bot, error) {
	if s.botErr != nil {
		return nil, s.botErr
	}
	return s.bot, nil
}

func (s *getBotRepositoryStub) BotsByOwnerID(context.Context, bots.UserID) ([]*bots.Bot, error) {
	return nil, nil
}

func (s *getBotRepositoryStub) SaveBot(context.Context, *bots.Bot) error {
	return nil
}

func (s *getBotRepositoryStub) UpdateBot(context.Context, *bots.Bot) error {
	return nil
}

func TestGetBotHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		repo := &getBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", false)}
		h := query.NewGetBotHandler(repo, logger)

		res, err := h.Handle(t.Context(), query.GetBotRequest{ActorID: 42, BotID: "b0001"})
		require.NoError(t, err)
		require.Equal(t, "b0001", res.ID)
		require.Equal(t, int64(42), res.OwnerID)
		require.Equal(t, "desc", res.Desc)
	})

	t.Run("deleted bot is hidden as not found", func(t *testing.T) {
		repo := &getBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 42, "sc0001", true)}
		h := query.NewGetBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetBotRequest{ActorID: 42, BotID: "b0001"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
	})

	t.Run("foreign bot is hidden as not found", func(t *testing.T) {
		repo := &getBotRepositoryStub{bot: testkit.MustValidBot(t, "b0001", 1, "sc0001", false)}
		h := query.NewGetBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetBotRequest{ActorID: 10, BotID: "b0001"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		repo := &getBotRepositoryStub{botErr: repoErr}
		h := query.NewGetBotHandler(repo, logger)

		_, err := h.Handle(t.Context(), query.GetBotRequest{ActorID: 10, BotID: "b0001"})
		require.ErrorIs(t, err, repoErr)
	})
}
