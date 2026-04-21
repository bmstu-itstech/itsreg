package telegram

import (
	"context"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type MessageSender struct {
	cl *http.Client
}

func NewMessageSender(cl *http.Client) *MessageSender {
	return &MessageSender{cl}
}

func (s *MessageSender) Send(
	_ context.Context, token bots.Token, userID bots.UserID, msg bots.BotMessage,
) error {
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
	if isTooManyRequests(err) {
		return port.ErrMessageSendRateLimitExceeded
	}
	if isForbiddenError(err) {
		return port.ErrUserBlockedBot
	}
	if err != nil {
		return err
	}

	return nil
}

func isForbiddenError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "forbidden")
}

func isTooManyRequests(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "too many requests")
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
