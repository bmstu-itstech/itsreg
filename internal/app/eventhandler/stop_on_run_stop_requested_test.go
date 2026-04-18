package eventhandler_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/eventhandler"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func TestStopOnRunStopRequestedHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("unexpected event type", func(t *testing.T) {
		rr := &runRepositoryStub{}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: "r1001", BotID: "b1001", Time: time.Now()})
		require.Error(t, err)
		require.Equal(t, 0, im.stopCalls)
	})

	t.Run("run repository error", func(t *testing.T) {
		rrErr := errors.New("db unavailable")
		rr := &runRepositoryStub{runErr: rrErr}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: "r1002", BotID: "b1002", Time: time.Now()})
		require.ErrorIs(t, err, rrErr)
		require.Equal(t, 0, im.stopCalls)
	})

	t.Run("instance manager stop fails and run transitions to failed", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r1003", "b1003", "token_b1003", bots.StatusStopping, &startedAt, nil)
		stopErr := errors.New("telegram unavailable")
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{stopErr: stopErr}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, stopErr)
		require.Equal(t, 1, im.stopCalls)
		require.Equal(t, run.BotID(), im.stoppedID)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Equal(t, bots.StatusFailed, run.Status())
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.failed", eb.published[0].EventName())
	})

	t.Run("instance manager stop fails and run fail transition is invalid", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		stoppedAt := startedAt.Add(10 * time.Minute)
		run := mustRestoreRun(t, "r1004", "b1004", "token_b1004", bots.StatusStopped, &startedAt, &stoppedAt)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{stopErr: errors.New("telegram unavailable")}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
		require.Equal(t, 1, im.stopCalls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("stop success updates run", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r1005", "b1005", "token_b1005", bots.StatusStopping, &startedAt, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.NoError(t, err)
		require.Equal(t, 1, im.stopCalls)
		require.Equal(t, run.BotID(), im.stoppedID)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Equal(t, bots.StatusStopped, run.Status())
		require.NotNil(t, run.StoppedAt())
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.stopped", eb.published[0].EventName())
	})

	t.Run("stop success but run state transition error", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		stoppedAt := startedAt.Add(10 * time.Minute)
		run := mustRestoreRun(t, "r1006", "b1006", "token_b1006", bots.StatusStopped, &startedAt, &stoppedAt)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
		require.Equal(t, 1, im.stopCalls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("update error on success path", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r1007", "b1007", "token_b1007", bots.StatusStopping, &startedAt, nil)
		updateErr := errors.New("update failed")
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}, updateErr: updateErr}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, updateErr)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("publish error on success path", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r1008", "b1008", "token_b1008", bots.StatusStopping, &startedAt, nil)
		publishErr := errors.New("publish failed")
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{publishErr: publishErr}
		h := eventhandler.NewStopOnRunStopRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStopRequested{RunID: run.ID(), BotID: run.BotID(), Time: time.Now()})
		require.ErrorIs(t, err, publishErr)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
	})
}
