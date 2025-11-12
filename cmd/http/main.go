package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	httpapi "github.com/bmstu-itstech/itsreg-bots/internal/api/http"
	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/command"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/query"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs"
	"github.com/bmstu-itstech/itsreg-bots/pkg/metrics"
	"github.com/bmstu-itstech/itsreg-bots/pkg/server"
)

func connectDB() (*sqlx.DB, error) {
	uri := os.Getenv("DATABASE_URI")
	if uri == "" {
		return nil, errors.New("DATABASE_URI must be set")
	}
	return sqlx.Connect("postgres", uri)
}

func main() {
	l := logs.DefaultLogger()
	mc := metrics.NoOp{}

	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}

	botRepository := service.NewPostgresBotRepository(db)
	participantRepository := service.NewPostgresParticipantRepository(db, l)
	usernameRepository := service.NewPostgresUsernameRepository(db)
	botMessageSender := service.NewTelegramMessageSender()

	process := ProcessHandlerAdapter{
		command.NewProcessHandler(botRepository, participantRepository, botMessageSender, l, mc),
	}
	entry := EntryHandlerAdapter{
		command.NewEntryHandler(botRepository, participantRepository, botMessageSender, l, mc),
	}

	telegramService := service.NewTelegramService(l, process, entry)

	a := app.Application{
		Commands: app.Commands{
			CreateBot: command.NewCreateBotHandler(botRepository, l, mc),
			Entry:     command.NewEntryHandler(botRepository, participantRepository, botMessageSender, l, mc),
			Process:   command.NewProcessHandler(botRepository, participantRepository, botMessageSender, l, mc),
			Start:     command.NewStartHandler(telegramService, botRepository, l, mc),
			Stop:      command.NewStopHandler(telegramService, l, mc),
			UpdateBot: command.NewUpdateBotHandler(botRepository, l, mc),
		},
		Queries: app.Queries{
			GetBot:      query.NewGetBotHandler(botRepository, l, mc),
			GetStatus:   query.NewGetStatusHandler(telegramService, l, mc),
			GetThreads:  query.NewGetThreadsHandler(participantRepository, usernameRepository, l, mc),
			GetUserBots: query.NewGetUserBotsHandler(botRepository, l, mc),
		},
	}

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return httpapi.HandlerFromMux(httpapi.NewHTTPServer(&a), router)
	})
}

// Страшно, очень страшно.
// Как сделать иначе?

type ProcessHandlerAdapter struct {
	H command.ProcessHandler
}

func (a ProcessHandlerAdapter) Process(
	ctx context.Context, botID bots.BotID, userID bots.UserID, msg bots.Message,
) error {
	return a.H.Handle(ctx, request.ProcessCommand{
		BotID:   string(botID),
		UserID:  int64(userID),
		Message: dto.Message{Text: msg.Text()},
	})
}

type EntryHandlerAdapter struct {
	H command.EntryHandler
}

func (a EntryHandlerAdapter) Entry(ctx context.Context, botID bots.BotID, userID bots.UserID, key bots.EntryKey) error {
	return a.H.Handle(ctx, request.EntryCommand{
		BotID:  string(botID),
		UserID: int64(userID),
		Key:    string(key),
	})
}
