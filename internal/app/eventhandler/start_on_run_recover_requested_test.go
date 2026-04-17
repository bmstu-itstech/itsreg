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

func TestStartOnRunRecoverRequestedHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("unexpected event type", func(t *testing.T) {
		rr := &runRepositoryStub{}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunRecoverRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: "r0001", BotID: "b0001", Time: time.Now()})
		require.Error(t, err)
		require.Equal(t, 0, im.startCalls)
	})

	t.Run("run repository error", func(t *testing.T) {
		rrErr := errors.New("db unavailable")
		rr := &runRepositoryStub{runErr: rrErr}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunRecoverRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunRecoverRequested{RunID: "r0001", Time: time.Now()})
		require.ErrorIs(t, err, rrErr)
		require.Equal(t, 0, im.startCalls)
	})

	t.Run("instance manager start fails and run transitions to failed", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r0010", "b0010", "token_b0010", bots.StatusActive, &startedAt, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{startErr: errors.New("telegram unavailable")}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunRecoverRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunRecoverRequested{RunID: run.ID(), Time: time.Now()})
		require.NoError(t, err)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, 1, rr.updateCalls)
		require.Equal(t, 1, eb.publishCalls)
		require.Equal(t, bots.StatusFailed, run.Status())
		require.Len(t, eb.published, 1)
		require.Equal(t, "run.failed", eb.published[0].EventName())
	})

	t.Run("instance manager start fails and run fail transition is invalid", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		stoppedAt := startedAt.Add(10 * time.Minute)
		run := mustRestoreRun(t, "r0011", "b0011", "token_b0011", bots.StatusStopped, &startedAt, &stoppedAt)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{startErr: errors.New("telegram unavailable")}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunRecoverRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunRecoverRequested{RunID: run.ID(), Time: time.Now()})
		require.ErrorIs(t, err, bots.ErrIllegalStateTransition)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("recover success does not update run", func(t *testing.T) {
		startedAt := time.Date(2026, time.April, 11, 12, 0, 0, 0, time.UTC)
		run := mustRestoreRun(t, "r0012", "b0012", "token_b0012", bots.StatusActive, &startedAt, nil)
		rr := &runRepositoryStub{runs: map[bots.RunID]*bots.Run{run.ID(): run}}
		im := &instanceManagerStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewStartOnRunRecoverRequestedHandler(rr, im, eb, logger)

		err := h.Handle(t.Context(), bots.RunRecoverRequested{RunID: run.ID(), Time: time.Now()})
		require.NoError(t, err)
		require.Equal(t, 1, im.startCalls)
		require.Equal(t, run.BotID(), im.startedID)
		require.Equal(t, run.Token(), im.startedToken)
		require.Equal(t, 0, rr.updateCalls)
		require.Equal(t, 0, eb.publishCalls)
	})
}
