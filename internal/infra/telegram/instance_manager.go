package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type InstanceManager struct {
	m       sync.Map // map[string]*botInstance
	l       *slog.Logger
	process port.ProcessHandler
	entry   port.EntryHandler
}

func NewInstanceManager(log *slog.Logger, process port.ProcessHandler, entry port.EntryHandler) *InstanceManager {
	return &InstanceManager{
		l:       log,
		process: process,
		entry:   entry,
	}
}

func (m *InstanceManager) Start(ctx context.Context, id bots.BotID, token bots.Token) error {
	const op = "InstanceManager.Start"
	l := m.l.With(
		slog.String("op", op),
		slog.String("bot_id", string(id)),
	)

	_, ok := m.m.Load(id)
	if ok {
		// Перезапускаем бота, если он уже запущен
		l.InfoContext(ctx, "bot already exists; stopping instance...")
		err := m.Stop(ctx, id)
		if err != nil {
			l.ErrorContext(ctx, "failed to stop previous instance while starting", slog.String("error", err.Error()))
		}
	}

	ins, err := startBotInstance(id, token, m.process, m.entry, m.l)
	m.m.Store(id, ins) // В любом случае сохраняем, чтобы иметь status = dead
	if err != nil {
		l.ErrorContext(ctx, "failed to start bot", slog.String("error", err.Error()))
		return fmt.Errorf("failed to start bot instance %s: %w", id, err)
	}
	l.InfoContext(ctx, "bot instance started")

	return nil
}

func (m *InstanceManager) Stop(_ context.Context, id bots.BotID) error {
	r, ok := m.m.Load(id)
	ins, _ := r.(*botInstance)
	if !ok {
		return fmt.Errorf("%w: %s", port.ErrRunningInstanceNotFound, id)
	}
	ins.Stop()
	m.m.Delete(id)
	return nil
}

type botInstance struct {
	botID   bots.BotID
	token   bots.Token
	api     *tgbotapi.BotAPI
	stopCh  chan struct{}
	process port.ProcessHandler
	entry   port.EntryHandler
	log     *slog.Logger
	dead    bool
}

func startBotInstance(
	botID bots.BotID,
	token bots.Token,
	process port.ProcessHandler,
	entry port.EntryHandler,
	log *slog.Logger,
) (*botInstance, error) {
	api, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		return nil, err
	}

	i := &botInstance{
		botID:   botID,
		token:   token,
		api:     api,
		stopCh:  make(chan struct{}),
		process: process,
		entry:   entry,
		log:     log,
		dead:    false,
	}

	conf := tgbotapi.NewUpdate(0)
	updates, err := api.GetUpdatesChan(conf)
	if err != nil {
		i.dead = true
		return nil, err
	}
	go i.run(updates)

	return i, nil
}

func (i *botInstance) IsDead() bool {
	return i.dead
}

func (i *botInstance) Stop() {
	i.dead = false
	i.stopCh <- struct{}{}
}

func (i *botInstance) run(updates tgbotapi.UpdatesChannel) {
	run := true
	for run {
		select {
		case update := <-updates:
			i.handleUpdate(context.Background(), update)
		case <-i.stopCh:
			run = false
		}
	}
	close(i.stopCh)
	i.api.StopReceivingUpdates()
}

func (i *botInstance) handleUpdate(ctx context.Context, upd tgbotapi.Update) {
	const op = "botInstance.handleUpdate"
	l := i.log.With(
		slog.String("op", op),
		slog.String("bot_id", string(i.botID)),
	)

	if upd.Message == nil {
		return
	}

	var err error
	if upd.Message.IsCommand() {
		err = i.entry.Entry(ctx, i.botID, bots.UserID(upd.Message.Chat.ID), bots.EntryKey(upd.Message.Command()))
	} else {
		if msg, err2 := bots.NewMessage(upd.Message.Text); err2 == nil {
			err = i.process.Process(ctx, i.botID, bots.UserID(upd.Message.Chat.ID), msg)
		} else {
			l.WarnContext(ctx, "unhandled message", slog.String("message", fmt.Sprintf("%v", upd.Message)))
		}
	}

	if err != nil {
		l.ErrorContext(ctx, "failed to handle update", slog.String("error", err.Error()))
	}
}
