package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (m *InstanceManager) Username(ctx context.Context, prtID bots.ParticipantID) (bots.Username, error) {
	const op = "InstanceManager.Username"
	l := m.l.With(
		slog.String("op", op),
	)

	r, ok := m.m.Load(prtID.BotID())
	if !ok {
		l.WarnContext(ctx, "instance is not running", slog.String("bot_id", string(prtID.BotID())))
		return "", fmt.Errorf("%w: %s", port.ErrRunningInstanceNotFound, prtID.BotID())
	}
	ins, _ := r.(*botInstance)

	uid := int64(prtID.UserID())
	chat, err := ins.api.GetChat(tgbotapi.ChatConfig{ChatID: uid})
	if err != nil {
		l.ErrorContext(ctx, "failed to get chat", slog.Int64("user_id", uid), slog.String("error", err.Error()))
		return "", err
	}

	if chat.UserName == "" {
		return "", port.ErrUsernameNotFound
	}

	return bots.Username(chat.UserName), nil
}
