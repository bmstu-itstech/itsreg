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
)

type mailingRepositoryStub struct {
	mailing            *bots.Mailing
	mailingErr         error
	mailingCalls       int
	requestedMailingID bots.MailingID
}

func (s *mailingRepositoryStub) Mailing(_ context.Context, id bots.MailingID) (*bots.Mailing, error) {
	s.mailingCalls++
	s.requestedMailingID = id
	if s.mailingErr != nil {
		return nil, s.mailingErr
	}
	return s.mailing, nil
}

func (s *mailingRepositoryStub) MailingsByOwnerID(
	context.Context,
	bots.UserID,
	port.MailingsFilter,
) ([]*bots.Mailing, error) {
	return nil, nil
}

func (s *mailingRepositoryStub) SaveMailing(context.Context, *bots.Mailing) error {
	return nil
}

func (s *mailingRepositoryStub) UpdateMailing(context.Context, *bots.Mailing) error {
	return nil
}

func TestSendMailingMessageHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mailingID := bots.MailingID("m2001")
	botID := bots.BotID("b2001")
	botToken := bots.Token("token_b2001")
	userID := bots.UserID(12345)
	msg := bots.MustNewMessage("hello").PromoteToBotMessage(nil)
	ev := bots.SendMailingMessageRequested{
		MailingID: mailingID,
		BotID:     botID,
		UserID:    userID,
		Message:   msg,
		Time:      time.Now(),
	}
	meta := dto.BotMeta{ID: string(botID), Token: string(botToken)}

	t.Run("unexpected event type", func(t *testing.T) {
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{}
		bmp := &botMetaProviderStub{}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: "r2001", BotID: "b2001", Time: time.Now()})
		require.Error(t, err)
		require.Equal(t, 0, mr.mailingCalls)
		require.Equal(t, 0, bmp.metaCalls)
		require.Equal(t, 0, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("mailing repository returns error", func(t *testing.T) {
		mrErr := errors.New("mailing repository unavailable")
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailingErr: mrErr}
		bmp := &botMetaProviderStub{}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, mrErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, mailingID, mr.requestedMailingID)
		require.Equal(t, 0, bmp.metaCalls)
		require.Equal(t, 0, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
	})

	t.Run("bot meta provider returns error", func(t *testing.T) {
		metaErr := errors.New("bot meta unavailable")
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{metaErr: metaErr}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, metaErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 0, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
	})

	t.Run("rate limiter returns error", func(t *testing.T) {
		waitErr := errors.New("rate limiter unavailable")
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{waitErr: waitErr}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, waitErr)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, botToken, rl.waitToken)
		require.Equal(t, 0, ms.sendCalls)
	})

	t.Run("context canceled while waiting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{waits: []time.Duration{time.Second}}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(ctx, ev)
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 1, mr.mailingCalls)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
	})

	t.Run("user blocked bot is ignored and counted as failed", func(t *testing.T) {
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{sendErr: port.ErrUserBlockedBot}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, botToken, ms.sentToken)
		require.Equal(t, userID, ms.sentUserID)
		require.Equal(t, msg, ms.sentMsg)
		require.Equal(t, 1, mailing.FailCount())
		require.Equal(t, 1, mailing.RecipientsTotal())
	})

	t.Run("send rate limit exceeded returns error and marks failed", func(t *testing.T) {
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{sendErr: port.ErrMessageSendRateLimitExceeded}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, port.ErrMessageSendRateLimitExceeded)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 1, mailing.FailCount())
		require.Equal(t, 1, mailing.RecipientsTotal())
	})

	t.Run("message sender generic error marks failed", func(t *testing.T) {
		sendErr := errors.New("network timeout")
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{sendErr: sendErr}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, sendErr)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 1, mailing.FailCount())
		require.Equal(t, 1, mailing.RecipientsTotal())
	})

	t.Run("send success", func(t *testing.T) {
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusStarted, []bots.UserID{userID})
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 1, mailing.SuccessCount())
		require.Equal(t, 0, mailing.FailCount())
		require.Equal(t, 1, mailing.RecipientsTotal())
	})

	t.Run("message sent state transition error is returned", func(t *testing.T) {
		mailing := mustRestoreMailing(t, mailingID, botID, bots.MailingStatusScheduled, []bots.UserID{userID})
		ms := &messageSenderStub{}
		mr := &mailingRepositoryStub{mailing: mailing}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendMailingMessageHandler(ms, mr, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, bots.ErrMailingIsNotStarted)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 0, mailing.SuccessCount())
		require.Equal(t, 1, mailing.RecipientsTotal())
	})
}

func mustRestoreMailing(
	t *testing.T,
	id bots.MailingID,
	botID bots.BotID,
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
		"start",
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
