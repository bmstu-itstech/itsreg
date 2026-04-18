package telegram

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type MessageSender struct {
	cl *http.Client
	l  *slog.Logger
}

func NewMessageSender(cl *http.Client, l *slog.Logger) *MessageSender {
	return &MessageSender{cl, l}
}

func (s *MessageSender) Send(
	ctx context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage,
) error {
	l := s.l.With(
		slog.String("op", "telegram.MessageSender.Send"),
		slog.Int64("user_id", userID.Int64()),
		slog.String("message", msg.String()),
	)

	api, err := tgbotapi.NewBotAPIWithClient(string(token), s.cl)
	if err != nil {
		return err
	}

	m := tgbotapi.NewMessage(int64(userID), msg.Text())
	m.ParseMode = tgbotapi.ModeHTML // Меньше головной боли с пользовательским вводом
	if opts := msg.Options(); len(opts) > 0 {
		m.ReplyMarkup = buildInlineKeyboardMarkup(opts)
	} else {
		m.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
	}

	_, err = api.Send(m)
	if isCantParseEntitiesError(err) {
		l.InfoContext(ctx, "can't parse HTML entities in message, send message without formatting",
			slog.String("error", err.Error()),
		)
		return err
	} else if isForbiddenError(err) {
		l.InfoContext(ctx, "user blocked bot, can't send message", slog.String("error", err.Error()))
		return port.ErrUserBlockedBot
	} else if err != nil {
		l.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func isCantParseEntitiesError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "can't parse entities")
}

func isForbiddenError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Forbidden")
}

func buildInlineKeyboardMarkup(opts []bots.Option) tgbotapi.ReplyKeyboardMarkup {
	rows := make([][]tgbotapi.KeyboardButton, len(opts))
	for i, opt := range opts {
		rows[i] = []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(opt.String()),
		}
	}
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}
