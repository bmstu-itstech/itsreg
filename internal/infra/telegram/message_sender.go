package telegram

import (
	"context"
	"log/slog"

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
	_ context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage,
) error {
	api, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		return err
	}

	m := tgbotapi.NewMessage(int64(userID), msg.Text())
	m.ParseMode = tgbotapi.ModeHTML
	if opts := msg.Options(); len(opts) > 0 {
		m.ReplyMarkup = buildInlineKeyboardMarkup(opts)
	}

	_, err = api.Send(m)
	return err
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
