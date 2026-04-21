package command_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type stopRunRepositoryStub struct {
	run            *bots.Run
	runErr         error
	updateErr      error
	runCalls       int
	updateCalls    int
	requestedRunID bots.RunID
	updated        *bots.Run
}

func (s *stopRunRepositoryStub) Run(_ context.Context, id bots.RunID) (*bots.Run, error) {
	s.runCalls++
	s.requestedRunID = id
	if s.runErr != nil {
		return nil, s.runErr
	}
	return s.run, nil
}

func (s *stopRunRepositoryStub) RunsByOwnerID(context.Context, bots.UserID, port.RunsFilter) ([]*bots.Run, error) {
	return nil, nil
}

func (s *stopRunRepositoryStub) ActiveRuns(context.Context) ([]*bots.Run, error) {
	return nil, nil
}

func (s *stopRunRepositoryStub) SaveRun(context.Context, *bots.Run) error {
	return nil
}

func (s *stopRunRepositoryStub) UpdateRun(_ context.Context, run *bots.Run) error {
	s.updateCalls++
	s.updated = run
	return s.updateErr
}

type stopRunBotMetaProviderStub struct {
	meta           dto.BotMeta
	metaErr        error
	calls          int
	requestedBotID bots.BotID
}

func (s *stopRunBotMetaProviderStub) BotMeta(_ context.Context, id bots.BotID) (dto.BotMeta, error) {
	s.calls++
	s.requestedBotID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

type stopRunEventBusStub struct {
	publishErr   error
	publishCalls int
	published    []event.Event
}

func (s *stopRunEventBusStub) Publish(_ context.Context, events ...event.Event) error {
	s.publishCalls++
	s.published = append(s.published, events...)
	return s.publishErr
}

func (s *stopRunEventBusStub) Subscribe(string, port.EventHandler) error {
	return nil
}

func TestStopRunHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0001",
			"b0001",
			"token_b0001",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0001", OwnerID: 42}}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		res, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0001"})
		require.NoError(t, err)
		require.Equal(t, command.StopRunResponse{}, res)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, bots.RunID("r0001"), rr.requestedRunID)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, bots.BotID("b0001"), bmp.requestedBotID)
		require.Equal(t, 1, rr.updateCalls)
		require.NotNil(t, rr.updated)
		require.Equal(t, bots.StatusStopping, rr.updated.Status())
		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.stop_requested", eb.published[0].EventName())
	})

	t.Run("run not found", func(t *testing.T) {
		rr := &stopRunRepositoryStub{runErr: port.ErrRunNotFound}
		bmp := &stopRunBotMetaProviderStub{}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0002"})
		require.ErrorIs(t, err, port.ErrRunNotFound)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 0, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("run repository error", func(t *testing.T) {
		rrErr := errors.New("db unavailable")
		rr := &stopRunRepositoryStub{runErr: rrErr}
		bmp := &stopRunBotMetaProviderStub{}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0003"})
		require.ErrorIs(t, err, rrErr)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 0, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot not found", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0004",
			"b0004",
			"token_b0004",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{metaErr: port.ErrBotNotFound}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0004"})
		require.ErrorIs(t, err, port.ErrBotNotFound)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot meta provider error", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0005",
			"b0005",
			"token_b0005",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		rr := &stopRunRepositoryStub{run: run}
		providerErr := errors.New("provider unavailable")
		bmp := &stopRunBotMetaProviderStub{metaErr: providerErr}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0005"})
		require.ErrorIs(t, err, providerErr)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("deleted bot is hidden as run not found", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0006",
			"b0006",
			"token_b0006",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0006", OwnerID: 42, Deleted: true}}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0006"})
		require.ErrorIs(t, err, port.ErrRunNotFound)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("foreign owner is hidden as run not found", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0007",
			"b0007",
			"token_b0007",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0007", OwnerID: 10}}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0007"})
		require.ErrorIs(t, err, port.ErrRunNotFound)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("illegal state transition", func(t *testing.T) {
		run := mustRestoreRun(t, "r0008", "b0008", "token_b0008", bots.StatusStarting, nil, nil)
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0008", OwnerID: 42}}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0008"})
		require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("update repository error", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0009",
			"b0009",
			"token_b0009",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		updateErr := errors.New("update failed")
		rr := &stopRunRepositoryStub{run: run, updateErr: updateErr}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0009", OwnerID: 42}}
		eb := &stopRunEventBusStub{}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0009"})
		require.ErrorIs(t, err, updateErr)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
		require.Equal(t, bots.StatusStopping, run.Status())
	})

	t.Run("publish error", func(t *testing.T) {
		run := mustRestoreRun(
			t,
			"r0010",
			"b0010",
			"token_b0010",
			bots.StatusActive,
			timePtr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			nil,
		)
		publishErr := errors.New("publish failed")
		rr := &stopRunRepositoryStub{run: run}
		bmp := &stopRunBotMetaProviderStub{meta: dto.BotMeta{ID: "b0010", OwnerID: 42}}
		eb := &stopRunEventBusStub{publishErr: publishErr}
		h := command.NewStopRunHandler(rr, bmp, eb, logger)

		_, err := h.Handle(t.Context(), command.StopRunRequest{ActorID: 42, RunID: "r0010"})
		require.ErrorIs(t, err, publishErr)
		require.Equal(t, 1, rr.runCalls)
		require.Equal(t, 1, bmp.calls)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.stop_requested", eb.published[0].EventName())
	})
}

func mustRestoreRun(
	t *testing.T,
	id string,
	botID string,
	token string,
	status bots.Status,
	startedAt *time.Time,
	stoppedAt *time.Time,
) *bots.Run {
	t.Helper()

	run, err := bots.RestoreRun(
		bots.RunID(id),
		bots.BotID(botID),
		bots.Token(token),
		status,
		nil,
		startedAt,
		stoppedAt,
	)
	require.NoError(t, err)
	return run
}

func timePtr(t time.Time) *time.Time {
	return &t
}
