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

func NewTelegramService(log *slog.Logger, process bots.ProcessHandler, entry bots.EntryHandler) *TelegramService {
	return &TelegramService{
		log:     log,
		process: process,
		entry:   entry,
	}
}

func (s *TelegramService) Status(_ context.Context, id bots.BotID) (bots.Status, error) {
	r, ok := s.m.Load(id)
	if !ok {
		return bots.Idle, nil
	}

	ins, _ := r.(*botInstance)
	if ins.IsDead() {
		return bots.Dead, nil
	}
	return bots.Running, nil
}

func (s *TelegramService) Start(ctx context.Context, id bots.BotID, token bots.Token) error {
	const op = "TelegramService.Start"
	l := s.log.With(
		slog.String("op", op),
		slog.String("bot_id", string(id)),
	)

	_, ok := s.m.Load(id)
	if ok {
		// Перезапускаем бота, если он уже запущен
		l.InfoContext(ctx, "bot already exists; stopping instance...")
		err := s.Stop(ctx, id)
		if err != nil {
			l.ErrorContext(ctx, "failed to stop previous instance while starting", slog.String("error", err.Error()))
		}
	}

	ins, err := startBotInstance(id, token, s.process, s.entry, s.log)
	s.m.Store(id, ins) // В любом случае сохраняем, чтобы иметь status = dead
	if err != nil {
		l.ErrorContext(ctx, "failed to start bot", slog.String("error", err.Error()))
		return fmt.Errorf("failed to start bot instance %s: %w", id, err)
	}

	return nil
}

func (s *TelegramService) Stop(_ context.Context, id bots.BotID) error {
	r, ok := s.m.Load(id)
	ins, _ := r.(*botInstance)
	if !ok {
		return fmt.Errorf("%w: %s", bots.ErrRunningInstanceNotFound, id)
	}
	ins.Stop()
	s.m.Delete(id)
	return nil
}

type botInstance struct {
	BotID   bots.BotID
	api     *tgbotapi.BotAPI
	stopCh  chan struct{}
	process bots.ProcessHandler
	entry   bots.EntryHandler
	log     *slog.Logger
	dead    bool
}

func startBotInstance(
	botID bots.BotID,
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
		BotID:   botID,
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
		slog.String("bot_id", string(i.BotID)),
	)

	if upd.Message == nil {
		return
	}

	var err error
	if upd.Message.IsCommand() {
		err = i.entry.Entry(ctx, i.BotID, bots.UserID(upd.Message.Chat.ID), bots.EntryKey(upd.Message.Command()))
	} else {
		if msg, err2 := bots.NewMessage(upd.Message.Text); err2 == nil {
			err = i.process.Process(ctx, i.BotID, bots.UserID(upd.Message.Chat.ID), msg)
		} else {
			l.WarnContext(ctx, "unhandled message", slog.String("message", fmt.Sprintf("%v", upd.Message)))
		}
	}

	if err != nil {
		l.ErrorContext(ctx, "failed to handle update", slog.String("error", err.Error()))
	}
}
