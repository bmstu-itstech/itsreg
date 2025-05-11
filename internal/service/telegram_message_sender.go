package service

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type TelegramMessageSender struct{}

func (s *TelegramMessageSender) Send(_ context.Context, token bots.Token, userId bots.UserId, msg bots.BotMessage) error {
	api, err := tg.NewBotAPI(string(token))
	if err != nil {
		return err
	}

	m := tg.NewMessage(int64(userId), msg.Text())
	if opts := msg.Options(); opts != nil {
		m.ReplyMarkup = buildInlineKeyboardMarkup(opts)
	}

	_, err = api.Send(m)
	return err
}

func buildInlineKeyboardMarkup(opts []bots.Option) tg.ReplyKeyboardMarkup {
	rows := make([][]tg.KeyboardButton, len(opts))
	for i, opt := range opts {
		rows[i] = []tg.KeyboardButton{
			tg.NewKeyboardButton(string(opt)),
		}
	}
	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}
