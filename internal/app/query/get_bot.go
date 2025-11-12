package query

import (
	"context"
	"log/slog"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/response"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/decorator"
)

type GetBotHandler decorator.QueryHandler[request.GetBotQuery, response.GetBotResponse]

type getBotHandler struct {
	bp port.BotProvider
}

func (h getBotHandler) Handle(ctx context.Context, q request.GetBotQuery) (response.GetBotResponse, error) {
	bot, err := h.bp.Bot(ctx, bots.BotID(q.ID))
	return dto.BotToDto(bot), err
}

func NewGetBotHandler(bp port.BotProvider, l *slog.Logger, mc decorator.MetricsClient) GetBotHandler {
	return decorator.ApplyQueryDecorators(getBotHandler{bp}, l, mc)
}
