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
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type createMailingRepositoryStub struct {
	saved       *bots.Mailing
	saveErr     error
	saveCalls   int
	updateCalls int
}

func (s *createMailingRepositoryStub) Mailing(context.Context, bots.MailingID) (*bots.Mailing, error) {
	return nil, nil
}

func (s *createMailingRepositoryStub) MailingsByOwnerID(
	context.Context, bots.UserID, port.MailingsFilter,
) ([]*bots.Mailing, error) {
	return nil, nil
}

func (s *createMailingRepositoryStub) SaveMailing(_ context.Context, mailing *bots.Mailing) error {
	s.saveCalls++
	s.saved = mailing
	return s.saveErr
}

func (s *createMailingRepositoryStub) UpdateMailing(context.Context, *bots.Mailing) error {
	s.updateCalls++
	return nil
}

type createMailingBotMetaProviderStub struct {
	meta           dto.BotMeta
	metaErr        error
	metaCalls      int
	requestedBotID bots.BotID
}

func (s *createMailingBotMetaProviderStub) BotMeta(_ context.Context, id bots.BotID) (dto.BotMeta, error) {
	s.metaCalls++
	s.requestedBotID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

type createMailingEventBusStub struct {
	publishErr   error
	publishCalls int
	published    []event.Event
}

func (s *createMailingEventBusStub) Publish(_ context.Context, events ...event.Event) error {
	s.publishCalls++
	s.published = append(s.published, events...)
	return s.publishErr
}

func (s *createMailingEventBusStub) Subscribe(string, port.EventHandler) error {
	return nil
}

func TestCreateMailingHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		res, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.NoError(t, err)
		require.NotEmpty(t, res.MailingID)

		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, bots.BotID("b0001"), bmp.requestedBotID)

		require.Equal(t, 1, mr.saveCalls)
		require.NotNil(t, mr.saved)
		require.Equal(t, res.MailingID, mr.saved.ID().String())
		require.Equal(t, bots.BotID("b0001"), mr.saved.BotID())
		require.Equal(t, bots.EntryKey("start"), mr.saved.EntryKey())
		require.ElementsMatch(t, []bots.UserID{1001, 1002}, mr.saved.Recipients())

		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.published, 1)
		require.Equal(t, "mailing.scheduled", eb.published[0].EventName())
	})

	t.Run("bot not found", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{metaErr: port.ErrBotNotFound}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 0, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot meta provider error", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmpErr := errors.New("db unavailable")
		bmp := &createMailingBotMetaProviderStub{metaErr: bmpErr}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, bmpErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 0, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot is deleted", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, true)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 0, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("permission denied for foreign bot owner", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(10, false)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 0, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("validation error on empty recipients", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.CreateMailingRequest{
			ActorID:    42,
			BotID:      "b0001",
			EntryKey:   "start",
			Recipients: nil,
		})
		var vErr shared.ValidationError
		require.ErrorAs(t, err, &vErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 0, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("mailing already exists is idempotent", func(t *testing.T) {
		mr := &createMailingRepositoryStub{saveErr: port.ErrMailingAlreadyExists}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		res, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.NoError(t, err)
		require.Empty(t, res.MailingID)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 1, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("save repository error", func(t *testing.T) {
		mrErr := errors.New("db unavailable")
		mr := &createMailingRepositoryStub{saveErr: mrErr}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createMailingEventBusStub{}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, mrErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 1, mr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("publish event error", func(t *testing.T) {
		mr := &createMailingRepositoryStub{}
		bmp := &createMailingBotMetaProviderStub{meta: validBotMeta(42, false)}
		ebErr := errors.New("event bus unavailable")
		eb := &createMailingEventBusStub{publishErr: ebErr}
		h := command.NewCreateMailingHandler(mr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateMailingRequest())
		require.ErrorIs(t, err, ebErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 1, mr.saveCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.published, 1)
		require.Equal(t, "mailing.scheduled", eb.published[0].EventName())
	})
}

func validCreateMailingRequest() command.CreateMailingRequest {
	return command.CreateMailingRequest{
		ActorID:    42,
		BotID:      "b0001",
		Name:       "Mailing",
		EntryKey:   "start",
		Recipients: []int64{1001, 1002},
	}
}
