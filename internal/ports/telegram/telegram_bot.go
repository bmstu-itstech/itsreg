package telegram

import (
	"context"
	"log/slog"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/command"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/query"
)

type telegramBot struct {
	botUUID string
	app     *app.Application
	log     *slog.Logger
	stopCh  chan struct{}
	api     *tg.BotAPI
}

func newTelegramBot(
	ctx context.Context,
	botUUID string,
	app *app.Application,
	log *slog.Logger,
) (*telegramBot, error) {
	appBot, err := app.Queries.GetBot.Handle(ctx, query.GetBot{BotUUID: botUUID})
	if err != nil {
		return nil, err
	}

	api, err := tg.NewBotAPI(appBot.Token)
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})

	return &telegramBot{
		botUUID: botUUID,
		app:     app,
		log:     log,
		stopCh:  stopCh,
		api:     api,
	}, nil
}

func (b *telegramBot) Start(ctx context.Context) error {
	conf := tg.NewUpdate(0)
	updates, err := b.api.GetUpdatesChan(conf)
	if err != nil {
		return err
	}

	err = b.app.Commands.UpdateStatus.Handle(ctx, command.UpdateStatus{
		BotUUID: b.botUUID,
		Status:  "started",
	})
	if err != nil {
		return err
	}

	go b.run(updates)

	return nil
}

func (b *telegramBot) Stop(ctx context.Context) error {
	err := b.app.Commands.UpdateStatus.Handle(ctx, command.UpdateStatus{
		BotUUID: b.botUUID,
		Status:  "stopped",
	})
	if err != nil {
		return err
	}

	b.stopCh <- struct{}{}

	return nil
}

func (b *telegramBot) SendMessage(toUserID int64, text string, buttons []string) error {
	msg := tg.NewMessage(toUserID, text)
	if len(buttons) > 0 {
		keyboard := buildInlineKeyboardMarkup(buttons)
		msg.ReplyMarkup = keyboard
	} else {
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	}
	msg.ParseMode = tg.ModeHTML

	_, err := b.api.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func buildInlineKeyboardMarkup(buttons []string) tg.ReplyKeyboardMarkup {
	rows := make([][]tg.KeyboardButton, len(buttons))
	for i, button := range buttons {
		rows[i] = []tg.KeyboardButton{tg.NewKeyboardButton(button)}
	}
	return tg.NewReplyKeyboard(rows...)
}

func (b *telegramBot) run(updates tg.UpdatesChannel) {
	run := true
	for run {
		select {
		case update := <-updates:
			b.handleUpdate(context.Background(), update)
		case <-b.stopCh:
			run = false
		}
	}
	close(b.stopCh)
	b.api.StopReceivingUpdates()
}

func (b *telegramBot) handleUpdate(ctx context.Context, update tg.Update) {
	var err error
	if update.Message != nil {
		if update.Message.IsCommand() {
			err = b.handleCommand(ctx, update.Message)
		} else {
			err = b.handleMessage(ctx, update.Message)
		}
	}

	if err != nil {
		b.log.Error("failed to handle update", "error", err.Error())
	}
}

func (b *telegramBot) handleCommand(ctx context.Context, msg *tg.Message) error {
	switch msg.Command() {
	case "start":
		return b.app.Commands.Entry.Handle(ctx, command.Entry{
			BotUUID: b.botUUID,
			UserID:  msg.Chat.ID,
			Key:     "start",
		})
	}
	return nil
}

func (b *telegramBot) handleMessage(ctx context.Context, msg *tg.Message) error {
	return b.app.Commands.Process.Handle(ctx, command.Process{
		BotUUID: b.botUUID,
		UserID:  msg.Chat.ID,
		Text:    msg.Text,
	})
}
