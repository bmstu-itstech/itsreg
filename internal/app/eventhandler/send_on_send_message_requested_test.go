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

type messageSenderStub struct {
	sendErr    error
	sendCalls  int
	sentToken  bots.Token
	sentUserID bots.UserID
	sentMsg    bots.BotMessage
}

func (s *messageSenderStub) Send(
	_ context.Context,
	token bots.Token,
	userID bots.UserID,
	msg bots.BotMessage,
) error {
	s.sendCalls++
	s.sentToken = token
	s.sentUserID = userID
	s.sentMsg = msg
	return s.sendErr
}

type rateLimiterStub struct {
	waitErr   error
	waitCalls int
	waitToken bots.Token
	waits     []time.Duration
}

func (s *rateLimiterStub) Wait(_ context.Context, token bots.Token, _ time.Time) (time.Duration, error) {
	s.waitCalls++
	s.waitToken = token

	if s.waitErr != nil {
		return 0, s.waitErr
	}

	if len(s.waits) == 0 {
		return 0, nil
	}

	wait := s.waits[0]
	s.waits = s.waits[1:]
	return wait, nil
}

type botMetaProviderStub struct {
	metaErr   error
	metaCalls int
	metaID    bots.BotID
	meta      dto.BotMeta
}

func (s *botMetaProviderStub) BotMeta(_ context.Context, id bots.BotID) (dto.BotMeta, error) {
	s.metaCalls++
	s.metaID = id
	if s.metaErr != nil {
		return dto.BotMeta{}, s.metaErr
	}
	return s.meta, nil
}

func TestSendOnSendMessageRequestedHandler_Handle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	botID := bots.BotID("b2001")
	botToken := bots.Token("token_b2001")
	msg := bots.MustNewMessage("hello").PromoteToBotMessage(nil)
	ev := bots.SendMessageRequested{
		BotID:   botID,
		UserID:  12345,
		Message: msg,
		Time:    time.Now(),
	}
	meta := dto.BotMeta{ID: string(botID), Token: string(botToken)}

	t.Run("unexpected event type", func(t *testing.T) {
		ms := &messageSenderStub{}
		bmp := &botMetaProviderStub{}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), bots.RunStartRequested{RunID: "r2001", BotID: "b2001", Time: time.Now()})
		require.Error(t, err)
		require.Equal(t, 0, bmp.metaCalls)
		require.Equal(t, 0, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("bot meta provider returns error", func(t *testing.T) {
		metaErr := errors.New("bot meta unavailable")
		ms := &messageSenderStub{}
		bmp := &botMetaProviderStub{metaErr: metaErr}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, metaErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 0, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("rate limiter returns error", func(t *testing.T) {
		waitErr := errors.New("rate limiter unavailable")
		ms := &messageSenderStub{}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{waitErr: waitErr}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, waitErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, botToken, rl.waitToken)
		require.Equal(t, 0, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("context canceled while waiting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		ms := &messageSenderStub{}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{waits: []time.Duration{time.Second}}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(ctx, ev)
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, 0, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("user blocked the bot is ignored", func(t *testing.T) {
		ms := &messageSenderStub{sendErr: port.ErrUserBlockedBot}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, botToken, ms.sentToken)
		require.Equal(t, ev.UserID, ms.sentUserID)
		require.Equal(t, ev.Message, ms.sentMsg)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("send rate limit exceeded returns error", func(t *testing.T) {
		ms := &messageSenderStub{sendErr: port.ErrMessageSendRateLimitExceeded}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, port.ErrMessageSendRateLimitExceeded)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("message sender generic error", func(t *testing.T) {
		sendErr := errors.New("network timeout")
		ms := &messageSenderStub{sendErr: sendErr}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.ErrorIs(t, err, sendErr)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, 0, eb.publishCalls)
	})

	t.Run("send success", func(t *testing.T) {
		ms := &messageSenderStub{}
		bmp := &botMetaProviderStub{meta: meta}
		rl := &rateLimiterStub{}
		eb := &eventBusStub{}
		h := eventhandler.NewSendOnSendMessageRequestedHandler(ms, bmp, rl, eb, logger)

		err := h.Handle(t.Context(), ev)
		require.NoError(t, err)
		require.Equal(t, 1, bmp.metaCalls)
		require.Equal(t, botID, bmp.metaID)
		require.Equal(t, 1, rl.waitCalls)
		require.Equal(t, botToken, rl.waitToken)
		require.Equal(t, 1, ms.sendCalls)
		require.Equal(t, botToken, ms.sentToken)
		require.Equal(t, ev.UserID, ms.sentUserID)
		require.Equal(t, ev.Message, ms.sentMsg)
		require.Equal(t, 0, eb.publishCalls)
	})
}
