package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type TelegramService struct {
	m       sync.Map // map[string]*botInstance
	log     *slog.Logger
	process bots.ProcessHandler
	entry   bots.EntryHandler
}

func (s *TelegramService) Status(_ context.Context, id bots.BotId) (bots.Status, error) {
	r, ok := s.m.Load(id)
	if !ok {
		return bots.Idle, nil
	}

	ins := r.(*botInstance)
	if ins.IsDead() {
		return bots.Dead, nil
	}
	return bots.Running, nil

}

func (s *TelegramService) Start(ctx context.Context, id bots.BotId, token bots.Token) error {
	_, ok := s.m.Load(id)
	if ok {
		// Перезапускаем бота, если он уже запущен
		s.log.Info("bot already exists; stopping instance...", "botId", id)
		err := s.Stop(ctx, id)
		if err != nil {
			s.log.Error("failed to stop previous instance while starting", "botId", id, "error", err.Error())
		}
	}

	ins, err := startBotInstance(id, token, s.process, s.entry, s.log)
	s.m.Store(id, ins) // В любом случае сохраняем, чтобы иметь status = dead
	if err != nil {
		return fmt.Errorf("failed to start bot instance %s: %w", id, err)
	}

	return nil
}

func (s *TelegramService) Stop(_ context.Context, id bots.BotId) error {
	r, ok := s.m.Load(id)
	ins := r.(*botInstance)
	if !ok {
		return fmt.Errorf("%w: %s", bots.ErrRunningInstanceNotFound, id)
	}
	ins.Stop()
	s.m.Delete(id)
	return nil
}

func NewTelegramService(log *slog.Logger, process bots.ProcessHandler, entry bots.EntryHandler) *TelegramService {
	return &TelegramService{
		log:     log,
		process: process,
		entry:   entry,
	}
}

type botInstance struct {
	botId   bots.BotId
	api     *tgbotapi.BotAPI
	stopCh  chan struct{}
	process bots.ProcessHandler
	entry   bots.EntryHandler
	log     *slog.Logger
	dead    bool
}

func startBotInstance(
	botId bots.BotId,
	token bots.Token,
	process bots.ProcessHandler,
	entry bots.EntryHandler,
	log *slog.Logger,
) (*botInstance, error) {
	api, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		return nil, err
	}

	i := &botInstance{
		botId:   botId,
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
	var err error
	if upd.Message != nil {
		if upd.Message.IsCommand() {
			err = i.entry.Entry(ctx, i.botId, bots.UserId(upd.Message.Chat.ID), bots.EntryKey(upd.Message.Command()))
		} else {
			if msg, err := bots.NewMessage(upd.Message.Text); err == nil {
				err = i.process.Process(ctx, i.botId, bots.UserId(upd.Message.Chat.ID), msg)
			} else {
				i.log.Warn("unhandled message", "message", upd.Message)
			}
		}
	}

	if err != nil {
		i.log.Error("failed to handle upd", "botId", i.botId, "error", err.Error())
	}
}
