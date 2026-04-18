package eventhandler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/eventhandler"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type runRepositoryStub struct {
	runs        map[bots.RunID]*bots.Run
	runErr      error
	updateErr   error
	updateCalls int
}

func (s *runRepositoryStub) Run(_ context.Context, id bots.RunID) (*bots.Run, error) {
	if s.runErr != nil {
		return nil, s.runErr
	}
	run, ok := s.runs[id]
	if !ok {
		return nil, port.ErrRunNotFound
	}
	return run, nil
}

func (s *runRepositoryStub) RunsByOwnerID(context.Context, bots.UserID, port.RunsFilter) ([]*bots.Run, error) {
	return nil, nil
}

func (s *runRepositoryStub) ActiveRuns(context.Context) ([]*bots.Run, error) {
	return nil, nil
}

func (s *runRepositoryStub) SaveRun(context.Context, *bots.Run) error {
	return nil
}

func (s *runRepositoryStub) UpdateRun(_ context.Context, _ *bots.Run) error {
	s.updateCalls++
	return s.updateErr
}

type instanceManagerStub struct {
	startErr     error
	startCalls   int
	startedID    bots.BotID
	startedToken bots.Token
}

func (s *instanceManagerStub) Start(_ context.Context, id bots.BotID, token bots.Token) error {
	s.startCalls++
	s.startedID = id
	s.startedToken = token
	return s.startErr
}

func (s *instanceManagerStub) Stop(context.Context, bots.BotID) error {
	return nil
}

type eventBusStub struct {
	publishErr   error
	publishCalls int
	published    []event.Event
}

func (s *eventBusStub) Publish(_ context.Context, events ...event.Event) error {
	s.publishCalls++
	s.published = append(s.published, events...)
	return s.publishErr
}

func (s *eventBusStub) Subscribe(string, port.EventHandler) error {
	return nil
}

func TestStartOnRunStartRequestedHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("unexpected event type", func(t *testing.T) {
		rr := &runRepositoryStub{}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunRecoverRequested{RunID: "r0001", Time: time.Now()})
		require.Error(t, err)
		require.Equal(t, 0, im.startCalls)
	})

	t.Run("run repository error", func(t *testing.T) {
		rrErr := errors.New("db unavailable")
		rr := &runRepositoryStub{runErr: rrErr}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: "r0001", BotID: "b0001", Time: time.Now()})
		require.ErrorIs(t, err, rrErr)
		require.Equal(t, 0, im.startCalls)
	})

	t.Run("instance manager start fails and run transitions to failed", func(t *testing.T) {
		run := mustRestoreRun(t, "r0001", "b0001", "token_b0001", bots.StatusStarting, nil, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{startErr: errors.New("telegram unavailable")}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.NoError(t, err)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Equal(t, bots.StatusFailed, run.Status())
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.failed", eb.published[0].EventName())
	})

	t.Run("start success updates run and publishes run.started", func(t *testing.T) {
		run := mustRestoreRun(t, "r0002", "b0002", "token_b0002", bots.StatusStarting, nil, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.NoError(t, err)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Equal(t, bots.StatusActive, run.Status())
		require.NotNil(t, run.StartedAt())
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.started", eb.published[0].EventName())
	})

	t.Run("start success but run state transition error", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r0003", "b0003", "token_b0003", bots.StatusActive, &startedAt, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("update error on success path", func(t *testing.T) {
		run := mustRestoreRun(t, "r0004", "b0004", "token_b0004", bots.StatusStarting, nil, nil)
		updateErr := errors.New("update failed")
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}, updateErr: updateErr}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, updateErr)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("publish error on success path", func(t *testing.T) {
		run := mustRestoreRun(t, "r0005", "b0005", "token_b0005", bots.StatusStarting, nil, nil)
		publishErr := errors.New("publish failed")
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{publishErr: publishErr}
		h := eventhandler.NewStartOnRunStartRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, publishErr)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
	})
}

func mustRestoreRun(
	t *testing.T,
	runID bots.RunID,
	botID bots.BotID,
	token bots.Token,
	status bots.Status,
	startedAt *time.Time,
	stoppedAt *time.Time,
) *bots.Run {
	t.Helper()
	run, err := bots.RestoreRun(runID, botID, token, status, nil, startedAt, stoppedAt)
	require.NoError(t, err)
	return run
}
