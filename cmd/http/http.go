package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	httpapi "github.com/bmstu-itstech/itsreg-bots/internal/api/http"
	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs"
	"github.com/bmstu-itstech/itsreg-bots/pkg/metrics"
	"github.com/bmstu-itstech/itsreg-bots/pkg/server"
)

func main() {
	l := logs.DefaultLogger()
	mc := metrics.NoOp{}

	botRepository := service.NewMockBotRepository()
	participantRepository := service.NewMockParticipantRepository()
	usernameRepository := service.NewMockUsernameRepository()
	botMessageSender := service.NewTelegramMessageSender()

	process := ProcessHandlerAdapter{app.NewProcessHandler(botRepository, participantRepository, botMessageSender, l, mc)}
	entry := EntryHandlerAdapter{app.NewEntryHandler(botRepository, participantRepository, botMessageSender, l, mc)}

	telegramService := service.NewTelegramService(l, process, entry)

	a := app.Application{
		Commands: app.Commands{
			CreateBot:      app.NewCreateBotHandler(botRepository, l, mc),
			Entry:          app.NewEntryHandler(botRepository, participantRepository, botMessageSender, l, mc),
			Process:        app.NewProcessHandler(botRepository, participantRepository, botMessageSender, l, mc),
			Start:          app.NewStartHandler(telegramService, botRepository, l, mc),
			Stop:           app.NewStopHandler(telegramService, l, mc),
			UpdateBot:      app.NewUpdateBotHandler(botRepository, l, mc),
			UpdateUsername: app.NewUpdateUsernameHandler(usernameRepository, l, mc),
		},
		Queries: app.Queries{
			GetBot:      app.NewGetBotHandler(botRepository, l, mc),
			GetStatus:   app.NewGetStatusHandler(telegramService, l, mc),
			GetThreads:  app.NewGetThreadsHandler(participantRepository, usernameRepository, l, mc),
			GetUserBots: app.NewGetUserBotsHandler(botRepository, l, mc),
			GetUsername: app.NewGetUsernameHandler(usernameRepository, l, mc),
		},
	}

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return httpapi.HandlerFromMux(httpapi.NewHTTPServer(&a), router)
	})
}

type ProcessHandlerAdapter struct {
	H app.ProcessHandler
}

func (a ProcessHandlerAdapter) Process(ctx context.Context, botId bots.BotId, userId bots.UserId, msg bots.Message) error {
	return a.H.Handle(ctx, app.Process{
		BotId:   string(botId),
		UserId:  int64(userId),
		Message: app.Message{Text: msg.Text()},
	})
}

type EntryHandlerAdapter struct {
	H app.EntryHandler
}

func (a EntryHandlerAdapter) Entry(ctx context.Context, botId bots.BotId, userId bots.UserId, key bots.EntryKey) error {
	return a.H.Handle(ctx, app.EntryCommand{
		BotId:  string(botId),
		UserId: int64(userId),
		Key:    string(key),
	})
}
