package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type UpdateUsername struct {
	UserID   int64
	Username string
}

type UpdateUsernameHandler decorator.CommandHandler[UpdateUsername]

type updateUsernameHandler struct {
	um bots.UsernameManager
}

func (h updateUsernameHandler) Handle(ctx context.Context, cmd UpdateUsername) error {
	return h.um.Upsert(ctx, bots.UserID(cmd.UserID), bots.Username(cmd.Username))
}

func NewUpdateUsernameHandler(
	um bots.UsernameManager,
	l *slog.Logger,
	mc decorator.MetricsClient,
) UpdateUsernameHandler {
	return decorator.ApplyCommandDecorators(updateUsernameHandler{um}, l, mc)
}
