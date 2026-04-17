package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type InstanceManager struct {
	m  sync.Map // map[string]*botInstance
	id port.InboundDispatcher
	l  *slog.Logger
}

func NewInstanceManager(id port.InboundDispatcher, l *slog.Logger) *InstanceManager {
	return &InstanceManager{
		id: id,
		l:  l,
	}
}

func (m *InstanceManager) Start(ctx context.Context, id bots.BotID, token bots.Token) error {
	l := m.l.With(
		slog.String("op", "telegram.InstanceManager.Start"),
		slog.String("bot_id", string(id)),
	)

	_, ok := m.m.Load(id)
	if ok {
		// Перезапускаем бота, если он уже запущен
		l.InfoContext(ctx, "bot already exists; stopping instance...")
		err := m.Stop(ctx, id)
		if err != nil {
			l.ErrorContext(ctx, "failed to stop previous instance while starting", slog.String("error", err.Error()))
			// Продолжаем попытку запустить нового бота
		}
	}

	ins, err := startBotInstance(id, token, m.id, m.l)
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
	botID  bots.BotID
	token  bots.Token
	api    *tgbotapi.BotAPI
	stopCh chan struct{}
	id     port.InboundDispatcher
	log    *slog.Logger
	dead   bool
}

func startBotInstance(
	botID bots.BotID,
	token bots.Token,
	id port.InboundDispatcher,
	log *slog.Logger,
) (*botInstance, error) {
	api, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		return nil, err
	}

	i := &botInstance{
		botID:  botID,
		token:  token,
		api:    api,
		stopCh: make(chan struct{}),
		id:     id,
		log:    log,
		dead:   false,
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
	const op = "telegram.botInstance.handleUpdate"
	l := i.log.With(
		slog.String("op", op),
		slog.String("bot_id", string(i.botID)),
	)

	if upd.Message == nil || upd.Message.From == nil {
		return
	}

	text := upd.Message.Text
	if upd.Message.IsCommand() {
		text = upd.Message.Command()
	}
	err := i.id.Dispatch(ctx, dto.InboundMessage{
		BotID:     i.botID.String(),
		UserID:    upd.Message.Chat.ID,
		Username:  usernameOrNil(*upd.Message.From),
		Text:      text,
		IsCommand: upd.Message.IsCommand(),
	})
	if err != nil {
		l.ErrorContext(ctx, "failed to handle update", slog.String("error", err.Error()))
	}
}

func usernameOrNil(user tgbotapi.User) *string {
	if user.UserName == "" {
		return nil
	}
	return &user.UserName
}
