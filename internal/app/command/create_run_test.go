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
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type createRunRepositoryStub struct {
	saved       *bots.Run
	saveErr     error
	saveCalls   int
	updateCalls int
}

func (s *createRunRepositoryStub) Run(context.Context, bots.RunID) (*bots.Run, error) {
	return nil, nil
}

func (s *createRunRepositoryStub) RunsByOwnerID(context.Context, bots.UserID, port.RunsFilter) ([]*bots.Run, error) {
	return nil, nil
}

func (s *createRunRepositoryStub) ActiveRuns(context.Context) ([]*bots.Run, error) {
	return nil, nil
}

func (s *createRunRepositoryStub) SaveRun(_ context.Context, run *bots.Run) error {
	s.saveCalls++
	s.saved = run
	return s.saveErr
}

func (s *createRunRepositoryStub) UpdateRun(context.Context, *bots.Run) error {
	s.updateCalls++
	return nil
}

type createRunBotMetaProviderStub struct {
	meta    dto.BotMeta
	metaErr error
}

func (s *createRunBotMetaProviderStub) BotMeta(context.Context, bots.BotID) (dto.BotMeta, error) {
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

type createRunEventBusStub struct {
	publishErr   error
	publishCalls int
	published    []event.Event
}

func (s *createRunEventBusStub) Publish(_ context.Context, events ...event.Event) error {
	s.publishCalls++
	s.published = append(s.published, events...)
	return s.publishErr
}

func (s *createRunEventBusStub) Subscribe(string, port.EventHandler) error {
	return nil
}

func TestCreateRunHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		res, err := h.Handle(t.Context(), validCreateRunRequest())
		require.NoError(t, err)
		require.NotEmpty(t, res.RunID)

		require.Equal(t, 1, rr.saveCalls)
		require.NotNil(t, rr.saved)
		require.Equal(t, bots.BotID("b0001"), rr.saved.BotID())
		require.Equal(t, bots.Token("token_b0001"), rr.saved.Token())
		require.Equal(t, bots.StatusStarting, rr.saved.Status())

		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.start_requested", eb.published[0].EventName())
	})

	t.Run("bot not found", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		bmp := &createRunBotMetaProviderStub{metaErr: port.ErrBotNotFound}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 0, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot meta provider error", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		dbErr := errors.New("db unavailable")
		bmp := &createRunBotMetaProviderStub{metaErr: dbErr}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, dbErr)
		require.Equal(t, 0, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot is deleted", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(42, true)}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 0, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("permission denied for foreign bot owner", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(10, false)}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 0, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("active run already exists", func(t *testing.T) {
		rr := &createRunRepositoryStub{saveErr: port.ErrActiveRunAlreadyExists}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, port.ErrActiveRunAlreadyExists)
		require.Equal(t, 1, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("save repository error", func(t *testing.T) {
		rrErr := errors.New("db unavailable")
		rr := &createRunRepositoryStub{saveErr: rrErr}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(42, false)}
		eb := &createRunEventBusStub{}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, rrErr)
		require.Equal(t, 1, rr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("publish event error", func(t *testing.T) {
		rr := &createRunRepositoryStub{}
		bmp := &createRunBotMetaProviderStub{meta: validBotMeta(42, false)}
		ebErr := errors.New("event bus unavailable")
		eb := &createRunEventBusStub{publishErr: ebErr}
		h := command.NewCreateRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), validCreateRunRequest())
		require.ErrorIs(t, err, ebErr)
		require.Equal(t, 1, rr.saveCalls)
		require.Equal(t, 1, eb.publishCalls)
	})
}

func validCreateRunRequest() command.CreateRunRequest {
	return command.CreateRunRequest{
		ActorID: 42,
		BotID:   "b0001",
	}
}

func validBotMeta(ownerID int64, deleted bool) dto.BotMeta {
	return dto.BotMeta{
		ID:      "b0001",
		OwnerID: ownerID,
		Token:   "token_b0001",
		Deleted: deleted,
	}
}
