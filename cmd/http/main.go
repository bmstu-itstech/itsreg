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
	"github.com/bmstu-itstech/itsreg-bots/internal/infra/postgres"
	"github.com/bmstu-itstech/itsreg-bots/internal/infra/telegram"
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

	repos := postgres.NewRepository(db, l)
	sender := telegram.NewMessageSender(l)

	process := ProcessHandlerAdapter{command.NewProcessHandler(repos, repos, sender, l, mc)}
	entry := EntryHandlerAdapter{command.NewEntryHandler(repos, repos, sender, l, mc)}
	instanceManager := telegram.NewInstanceManager(l, process, entry)

	a := app.Application{
		Commands: app.Commands{
			CreateBot: command.NewCreateBotHandler(repos, l, mc),
			Entry:     command.NewEntryHandler(repos, repos, sender, l, mc),
			Mailing:   command.NewMailingHandler(repos, repos, sender, l, mc),
			Process:   command.NewProcessHandler(repos, repos, sender, l, mc),
			Start:     command.NewStartHandler(instanceManager, repos, l, mc),
			Stop:      command.NewStopHandler(instanceManager, l, mc),
			UpdateBot: command.NewUpdateBotHandler(repos, l, mc),
		},
		Queries: app.Queries{
			GetBot:      query.NewGetBotHandler(repos, l, mc),
			GetStatus:   query.NewGetStatusHandler(instanceManager, repos, l, mc),
			GetThreads:  query.NewGetThreadsHandler(repos, instanceManager, l, mc),
			GetUserBots: query.NewGetUserBotsHandler(repos, l, mc),
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
