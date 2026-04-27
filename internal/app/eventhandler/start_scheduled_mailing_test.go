package eventhandler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/eventhandler"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared/event"
)

type startScheduledMailingRepositoryStub struct {
	mailing            *bots.Mailing
	mailingErr         error
	mailingCalls       int
	requestedMailingID bots.MailingID

	updateErr      error
	updateCalls    int
	updatedMailing *bots.Mailing
}

func (s *startScheduledMailingRepositoryStub) Mailing(
	_ context.Context,
	id bots.MailingID,
) (*bots.Mailing, error) {
	s.mailingCalls++
	s.requestedMailingID = id
	if s.mailingErr != nil {
		return nil, s.mailingErr
	}
	return s.mailing, nil
}

func (s *startScheduledMailingRepositoryStub) MailingsByOwnerID(
	context.Context,
	bots.UserID,
	port.MailingsFilter,
) ([]*bots.Mailing, error) {
	return nil, nil
}

func (s *startScheduledMailingRepositoryStub) SaveMailing(context.Context, *bots.Mailing) error {
	return nil
}

func (s *startScheduledMailingRepositoryStub) UpdateMailing(
	_ context.Context,
	mailing *bots.Mailing,
) error {
	s.updateCalls++
	s.updatedMailing = mailing
	return s.updateErr
}

type startScheduledThreadRepositoryStub struct {
	saveErr      error
	saveCalls    int
	savedThreads []*bots.Thread
}

func (s *startScheduledThreadRepositoryStub) LastUserThread(
	context.Context,
	bots.BotID,
	bots.UserID,
) (*bots.Thread, error) {
	return nil, nil
}

func (s *startScheduledThreadRepositoryStub) SaveThread(
	_ context.Context,
	thread *bots.Thread,
) error {
	s.saveCalls++
	s.savedThreads = append(s.savedThreads, thread)
	return s.saveErr
}

func (s *startScheduledThreadRepositoryStub) UpdateThread(context.Context, *bots.Thread) error {
	return nil
}

type startScheduledBotMetaProviderStub struct {
	meta      dto.BotMeta
	metaErr   error
	metaCalls int
	metaID    bots.BotID
}

func (s *startScheduledBotMetaProviderStub) BotMeta(
	_ context.Context,
	id bots.BotID,
) (dto.BotMeta, error) {
	s.metaCalls++
	s.metaID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

type startScheduledScriptRepositoryStub struct {
	script            *bots.Script
	scriptErr         error
	scriptCalls       int
	requestedScriptID bots.ScriptID
}

func (s *startScheduledScriptRepositoryStub) Script(
	_ context.Context,
	id bots.ScriptID,
) (*bots.Script, error) {
	s.scriptCalls++
	s.requestedScriptID = id
	if s.scriptErr != nil {
		return nil, s.scriptErr
	}
	return s.script, nil
}

func (s *startScheduledScriptRepositoryStub) ScriptsByOwnerID(
	context.Context,
	bots.UserID,
) ([]*bots.Script, error) {
	return nil, nil
}

func (s *startScheduledScriptRepositoryStub) SaveScript(context.Context, *bots.Script) error {
	return nil
}

func (s *startScheduledScriptRepositoryStub) UpdateScript(context.Context, *bots.Script) error {
	return nil
}

type startScheduledEventBusStub struct {
	publishErr       error
	publishCalls     int
	publishedBatches [][]event.Event
}

func (s *startScheduledEventBusStub) Publish(_ context.Context, events ...event.Event) error {
	s.publishCalls++
	batch := append([]event.Event(nil), events...)
	s.publishedBatches = append(s.publishedBatches, batch)
	return s.publishErr
}

func (s *startScheduledEventBusStub) Subscribe(string, port.EventHandler) error {
	return nil
}

func TestStartScheduledMailingHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mailingID := bots.MailingID("m3001")
	botID := bots.BotID("b3001")
	userID := bots.UserID(3001)
	ownerID := bots.UserID(9001)
	meta := dto.BotMeta{
		ID:       string(botID),
		OwnerID:  ownerID.Int64(),
		ScriptID: "script-3001",
		Token:    "token_b3001",
	}

	t.Run("unexpected event type", func(t *testing.T) {
		mr := &startScheduledMailingRepositoryStub{}
		tr := &startScheduledThreadRepositoryStub{}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		sr := &startScheduledScriptRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{
			RunID: "r3001",
			BotID: botID,
			Time:  time.Now(),
		})
		require.Error(t, err)
		require.Equal(t, 0, mr.mailingCalls)
		require.Equal(t, 0, bmp.metaCalls)
		require.Equal(t, 0, sr.scriptCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("mailing repository returns error", func(t *testing.T) {
		mailingErr := errors.New("mailing repository unavailable")
		mr := &startScheduledMailingRepositoryStub{mailingErr: mailingErr}
		tr := &startScheduledThreadRepositoryStub{}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		sr := &startScheduledScriptRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, mailingErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, mailingID, mr.requestedMailingID)
		require.Equal(t, 0, bmp.metaCalls)
		require.Equal(t, 0, sr.scriptCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot meta provider returns error", func(t *testing.T) {
		metaErr := errors.New("bot meta unavailable")
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{metaErr: metaErr}
		tr := &startScheduledThreadRepositoryStub{}
		sr := &startScheduledScriptRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, metaErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 0, sr.scriptCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("script repository returns error", func(t *testing.T) {
		scriptErr := errors.New("script repository unavailable")
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		sr := &startScheduledScriptRepositoryStub{scriptErr: scriptErr}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, scriptErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, sr.scriptCalls)
		require.Equal(t, bots.ScriptID(meta.ScriptID), sr.requestedScriptID)
		require.Equal(t, 0, mr.updateCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("mailing state transition error is returned", func(t *testing.T) {
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusStarted,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, _ := buildScheduledMailingScript(t, ownerID, bots.EntryKey("start"), "hello")
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, shared.ErrIllegalStateTransition)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 1, sr.scriptCalls)
		require.Equal(t, 0, mr.updateCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("update error on start path", func(t *testing.T) {
		updateErr := errors.New("update failed")
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing, updateErr: updateErr}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, _ := buildScheduledMailingScript(t, ownerID, bots.EntryKey("start"), "hello")
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, updateErr)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, bots.MailingStatusStarted, mr.updatedMailing.Status())
		require.NotNil(t, mr.updatedMailing.StartedAt())
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("script entry error stops recipient processing", func(t *testing.T) {
		mailing := mustRestoreMailingWithEntryKey(
			t,
			mailingID,
			botID,
			"missing",
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, _ := buildScheduledMailingScript(t, ownerID, bots.EntryKey("start"), "hello")
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, bots.ErrEntryNotFound)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, 0, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("save thread error stops recipient processing", func(t *testing.T) {
		saveErr := errors.New("thread repository unavailable")
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, _ := buildScheduledMailingScript(t, ownerID, bots.EntryKey("start"), "hello")
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{saveErr: saveErr}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, saveErr)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, 1, tr.saveCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("publish error stops recipient processing", func(t *testing.T) {
		publishErr := errors.New("event bus unavailable")
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, _ := buildScheduledMailingScript(t, ownerID, bots.EntryKey("start"), "hello")
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{publishErr: publishErr}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, publishErr)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, 1, tr.saveCalls)
		require.Equal(t, 1, eb.publishCalls)
	})

	t.Run("success for one recipient publishes all messages", func(t *testing.T) {
		mailing := mustRestoreMailing(
			t,
			mailingID,
			botID,
			bots.MailingStatusScheduled,
			[]bots.UserID{userID},
		)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, expectedMessages := buildScheduledMailingScript(
			t,
			ownerID,
			bots.EntryKey("start"),
			"hello",
			"world",
		)
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, bots.MailingStatusStarted, mr.updatedMailing.Status())
		require.NotNil(t, mr.updatedMailing.StartedAt())
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, sr.scriptCalls)
		require.Equal(t, bots.ScriptID(meta.ScriptID), sr.requestedScriptID)
		require.Equal(t, 1, tr.saveCalls)
		require.Len(t, tr.savedThreads, 1)
		require.Equal(t, botID, tr.savedThreads[0].BotID())
		require.Equal(t, userID, tr.savedThreads[0].UserID())
		require.Equal(t, 1, eb.publishCalls)
		require.Len(t, eb.publishedBatches, 1)
		requirePublishedBatch(t, eb.publishedBatches[0], mailingID, botID, userID, expectedMessages)
	})

	t.Run("success for multiple recipients batches messages per recipient", func(t *testing.T) {
		recipients := []bots.UserID{bots.UserID(3002), bots.UserID(3003)}
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusScheduled, recipients)
		mr := &startScheduledMailingRepositoryStub{mailing: mailing}
		bmp := &startScheduledBotMetaProviderStub{meta: meta}
		script, expectedMessages := buildScheduledMailingScript(
			t,
			ownerID,
			bots.EntryKey("start"),
			"hello",
			"world",
		)
		sr := &startScheduledScriptRepositoryStub{script: script}
		tr := &startScheduledThreadRepositoryStub{}
		eb := &startScheduledEventBusStub{}
		h := eventhandler.NewStartScheduledMailing(mr, tr, bmp, sr, eb, logger)

		ev := bots.MailingScheduled{MailingID: mailingID, Time: time.Now()}
		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, mr.updateCalls)
		require.Equal(t, bots.MailingStatusStarted, mr.updatedMailing.Status())
		require.Equal(t, len(recipients), tr.saveCalls)
		require.Equal(t, len(recipients), eb.publishCalls)
		require.Len(t, tr.savedThreads, len(recipients))
		require.Len(t, eb.publishedBatches, len(recipients))

		savedUsers := make(map[bots.UserID]struct{}, len(recipients))
		for _, thread := range tr.savedThreads {
			savedUsers[thread.UserID()] = struct{}{}
			require.Equal(t, botID, thread.BotID())
		}
		require.Len(t, savedUsers, len(recipients))
		for _, recipient := range recipients {
			_, ok := savedUsers[recipient]
			require.True(t, ok, "missing saved thread for recipient %d", recipient)
		}

		publishedByUser := make(map[bots.UserID][]event.Event, len(recipients))
		for _, batch := range eb.publishedBatches {
			requirePublishedBatchHeader(t, batch, mailingID, botID, expectedMessages)
			ev := batch[0].(bots.SendMailingMessageRequested)
			publishedByUser[ev.UserID] = batch
		}
		require.Len(t, publishedByUser, len(recipients))
		for _, recipient := range recipients {
			batch, ok := publishedByUser[recipient]
			require.True(t, ok, "missing published batch for recipient %d", recipient)
			requirePublishedBatch(t, batch, mailingID, botID, recipient, expectedMessages)
		}
	})
}

func buildScheduledMailingScript(
	t *testing.T,
	ownerID bots.UserID,
	entryKey bots.EntryKey,
	messageTexts ...string,
) (*bots.Script, []bots.BotMessage) {
	t.Helper()

	msgs := make([]bots.Message, len(messageTexts))
	expected := make([]bots.BotMessage, len(messageTexts))
	for i, text := range messageTexts {
		msg := bots.MustNewMessage(text)
		msgs[i] = msg
		if i == len(messageTexts)-1 {
			expected[i] = msg.PromoteToBotMessage([]bots.Option{})
			continue
		}
		expected[i] = msg.PromoteToBotMessage(nil)
	}

	node := bots.MustNewNode(bots.MustNewState(1), "scheduled-mailing", nil, msgs, nil)
	entry := bots.MustNewEntry(entryKey, bots.MustNewState(1))

	return bots.MustNewScript(
		ownerID,
		"scheduled mailing script",
		[]bots.Node{node},
		[]bots.Entry{entry},
	), expected
}

func mustRestoreMailingWithEntryKey(
	t *testing.T,
	id bots.MailingID,
	botID bots.BotID,
	entryKey bots.EntryKey,
	status bots.MailingStatus,
	recipients []bots.UserID,
) *bots.Mailing {
	t.Helper()

	createdAt := time.Now().Add(-time.Minute)
	var startedAt *time.Time
	if status == bots.MailingStatusStarted || status == bots.MailingStatusCompleted ||
		status == bots.MailingStatusFailed {
		tm := time.Now().Add(-30 * time.Second)
		startedAt = &tm
	}

	m, err := bots.RestoreMailing(
		id,
		botID,
		"Mailing",
		entryKey,
		status,
		recipients,
		nil,
		createdAt,
		startedAt,
		nil,
	)
	require.NoError(t, err)
	return m
}

func requirePublishedBatchHeader(
	t *testing.T,
	batch []event.Event,
	mailingID bots.MailingID,
	botID bots.BotID,
	expectedMessages []bots.BotMessage,
) {
	t.Helper()
	require.Len(t, batch, len(expectedMessages))
	for i, raw := range batch {
		ev, ok := raw.(bots.SendMailingMessageRequested)
		require.True(t, ok)
		require.Equal(t, mailingID, ev.MailingID)
		require.Equal(t, botID, ev.BotID)
		require.Equal(t, expectedMessages[i], ev.Message)
		require.False(t, ev.Time.IsZero())
	}
}

func requirePublishedBatch(
	t *testing.T,
	batch []event.Event,
	mailingID bots.MailingID,
	botID bots.BotID,
	userID bots.UserID,
	expectedMessages []bots.BotMessage,
) {
	t.Helper()
	requirePublishedBatchHeader(t, batch, mailingID, botID, expectedMessages)
	for _, raw := range batch {
		ev := raw.(bots.SendMailingMessageRequested)
		require.Equal(t, userID, ev.UserID)
	}
}
