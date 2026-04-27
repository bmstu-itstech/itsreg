package query_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type getBotMailingsRepositoryStub struct {
	mailings    []*bots.Mailing
	mailingsErr error
	calls       int
	gotOwnerID  bots.UserID
	gotFilter   port.MailingsFilter
}

func (s *getBotMailingsRepositoryStub) Mailing(context.Context, bots.MailingID) (*bots.Mailing, error) {
	return nil, nil
}

func (s *getBotMailingsRepositoryStub) MailingsByOwnerID(
	_ context.Context,
	ownerID bots.UserID,
	filter port.MailingsFilter,
) ([]*bots.Mailing, error) {
	s.calls++
	s.gotOwnerID = ownerID
	s.gotFilter = filter
	if s.mailingsErr != nil {
		return nil, s.mailingsErr
	}
	return s.mailings, nil
}

func (s *getBotMailingsRepositoryStub) SaveMailing(context.Context, *bots.Mailing) error {
	return nil
}

func (s *getBotMailingsRepositoryStub) UpdateMailing(context.Context, *bots.Mailing) error {
	return nil
}

type getBotMailingsMetaProviderStub struct {
	meta     dto.BotMeta
	metaErr  error
	calls    int
	gotBotID bots.BotID
}

func (s *getBotMailingsMetaProviderStub) BotMeta(_ context.Context, id bots.BotID) (dto.BotMeta, error) {
	s.calls++
	s.gotBotID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

func TestGetBotMailingsHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success maps mailings and passes bot id and status filters", func(t *testing.T) {
		status := "started"
		repo := &getBotMailingsRepositoryStub{
			mailings: []*bots.Mailing{
				mustRestoreMailing(t, "m0001", "b0001", bots.MailingStatusStarted),
				mustRestoreMailing(t, "m0002", "b0001", bots.MailingStatusFailed),
			},
		}
		bmp := &getBotMailingsMetaProviderStub{meta: dto.BotMeta{ID: "b0001"}}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		res, err := h.Handle(t.Context(), query.GetBotMailingsRequest{ActorID: 10, BotID: "b0001", Status: &status})
		require.NoError(t, err)

		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0001"), bmp.gotBotID)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(10), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.MailingStatusStarted, *repo.gotFilter.Status)

		require.Len(t, res, 2)
		require.Equal(t, "m0001", res[0].ID)
		require.Equal(t, "b0001", res[0].BotID)
		require.Equal(t, "started", res[0].Status)
		require.Equal(t, "m0002", res[1].ID)
		require.Equal(t, "failed", res[1].Status)
	})

	t.Run("without status passes bot id only", func(t *testing.T) {
		repo := &getBotMailingsRepositoryStub{
			mailings: []*bots.Mailing{mustRestoreMailing(t, "m0003", "b0003", bots.MailingStatusScheduled)},
		}
		bmp := &getBotMailingsMetaProviderStub{meta: dto.BotMeta{ID: "b0003"}}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		res, err := h.Handle(t.Context(), query.GetBotMailingsRequest{ActorID: 77, BotID: "b0003"})
		require.NoError(t, err)

		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0003"), bmp.gotBotID)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(77), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0003"), *repo.gotFilter.BotID)
		require.Nil(t, repo.gotFilter.Status)
		require.Len(t, res, 1)
		require.Equal(t, "m0003", res[0].ID)
	})

	t.Run("invalid status returns empty response and skips providers", func(t *testing.T) {
		invalidStatus := "not-a-status"
		repo := &getBotMailingsRepositoryStub{}
		bmp := &getBotMailingsMetaProviderStub{}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		res, err := h.Handle(
			t.Context(),
			query.GetBotMailingsRequest{ActorID: 10, BotID: "b0001", Status: &invalidStatus},
		)
		require.NoError(t, err)
		require.Empty(t, res)
		require.Equal(t, 0, bmp.calls)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("bot not found is returned and repository is not called", func(t *testing.T) {
		repo := &getBotMailingsRepositoryStub{}
		bmp := &getBotMailingsMetaProviderStub{metaErr: port.ErrBotNotFound}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotMailingsRequest{ActorID: 10, BotID: "b0099"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0099"), bmp.gotBotID)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("bot provider error is returned", func(t *testing.T) {
		providerErr := errors.New("provider unavailable")
		repo := &getBotMailingsRepositoryStub{}
		bmp := &getBotMailingsMetaProviderStub{metaErr: providerErr}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotMailingsRequest{ActorID: 10, BotID: "b0001"})
		require.ErrorIs(t, err, providerErr)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, repo.calls)
	})

	t.Run("mailings repository error is returned", func(t *testing.T) {
		status := "failed"
		repoErr := errors.New("db unavailable")
		repo := &getBotMailingsRepositoryStub{mailingsErr: repoErr}
		bmp := &getBotMailingsMetaProviderStub{meta: dto.BotMeta{ID: "b0001"}}
		h := query.NewGetBotMailingsHandler(repo, bmp, logger)

		_, err := h.Handle(t.Context(), query.GetBotMailingsRequest{ActorID: 11, BotID: "b0001", Status: &status})
		require.ErrorIs(t, err, repoErr)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 1, repo.calls)
		require.Equal(t, bots.UserID(11), repo.gotOwnerID)
		require.NotNil(t, repo.gotFilter.BotID)
		require.Equal(t, bots.BotID("b0001"), *repo.gotFilter.BotID)
		require.NotNil(t, repo.gotFilter.Status)
		require.Equal(t, bots.MailingStatusFailed, *repo.gotFilter.Status)
	})
}
