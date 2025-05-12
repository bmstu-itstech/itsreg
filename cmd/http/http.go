package main

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs"
	"github.com/bmstu-itstech/itsreg-bots/pkg/metrics"
)

func main() {
	l := logs.DefaultLogger()
	mc := metrics.NoOp{}

	botRepository := service.NewMockBotRepository()
	participantRepository := service.NewMockParticipantRepository()
	usernameRepository := service.NewMockUsernameRepository()
	botMessageSender := service.NewTelegramMessageSender()
	telegramService := service.NewTelegramService(l)

	_ = app.Application{
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
			GetThreads:  app.NewGetThreadsHandler(participantRepository, l, mc),
			GetUserBots: app.NewGetUserBotsHandler(botRepository, l, mc),
			GetUsername: app.NewGetUsernameHandler(usernameRepository, l, mc),
		},
	}
}
