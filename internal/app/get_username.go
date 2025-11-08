package app

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetUsername struct {
	UserID int64
}

type GetUsernameHandler decorator.QueryHandler[GetUsername, string]

type getUsernameHandler struct {
	up bots.UsernameProvider
}

func (h getUsernameHandler) Handle(ctx context.Context, q GetUsername) (string, error) {
	username, err := h.up.Username(ctx, bots.UserID(q.UserID))
	if err != nil {
		return "", err
	}
	return string(username), nil
}

func NewGetUsernameHandler(up bots.UsernameProvider, l *slog.Logger, mc decorator.MetricsClient) GetUsernameHandler {
	return decorator.ApplyQueryDecorators(getUsernameHandler{up}, l, mc)
}
