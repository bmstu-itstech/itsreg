package telegram

import (
	"context"
	"fmt"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type MessageSender struct {
	l *slog.Logger
}

func NewMessageSender(l *slog.Logger) *MessageSender {
	return &MessageSender{
		l: l,
	}
}

func (s *MessageSender) Send(
	ctx context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage,
) error {
	const op = "MessageSender.Send"
	l := s.l.With(
		slog.String("op", op),
		slog.String("message", msg.String()),
	)

	api, err := tgbotapi.NewBotAPI(string(token))
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
	if err != nil {
		if isCantParseEntitiesError(err) {
			l.WarnContext(ctx, "can't parse MarkdownV2 entities in message, send message without formatting",
				slog.String("error", err.Error()),
			)
			m.ParseMode = ""
			_, err = api.Send(m)
		} else if isForbiddenError(err) {
			l.WarnContext(ctx, "user blocked bot, can't send message",
				slog.String("error", err.Error()),
			)
			err = fmt.Errorf("%w: %d", port.ErrUserBlockedBot, userID)
		}
	}

	return err
}

func isCantParseEntitiesError(err error) bool {
	return strings.Contains(err.Error(), "can't parse entities")
}

func isForbiddenError(err error) bool {
	return strings.Contains(err.Error(), "Forbidden")
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
