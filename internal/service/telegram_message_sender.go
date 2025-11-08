package service

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type TelegramMessageSender struct{}

func NewTelegramMessageSender() *TelegramMessageSender {
	return &TelegramMessageSender{}
}

func (s *TelegramMessageSender) Send(
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
			tgbotapi.NewKeyboardButton(string(opt)),
		}
	}
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}
