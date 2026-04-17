package app

import (
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/eventhandler"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
)

type Commands struct {
	CreateBot    *command.CreateBotHandler
	CreateRun    *command.CreateRunHandler
	CreateScript *command.CreateScriptHandler
	DeleteBot    *command.DeleteBotHandler
	DeleteScript *command.DeleteScriptHandler
	Entry        *command.EntryHandler
	Process      *command.ProcessHandler
	UpdateBot    *command.UpdateBotHandler
	UpdateScript *command.UpdateScriptHandler
}

type Queries struct {
	GetBot     *query.GetBotHandler
	GetBots    *query.GetBotsHandler
	GetScript  *query.GetScriptHandler
	GetScripts *query.GetScriptsHandler
}

type EventHandlers struct {
	StartOnRunStartRequested   *eventhandler.StartOnRunStartRequestedHandler
	StartOnRunRecoverRequested *eventhandler.StartOnRunRecoverRequestedHandler
}

type Application struct {
	Commands Commands
	Queries  Queries
	Events   EventHandlers
}

type Infra struct {
	BotMetaProvider      port.BotMetaProvider
	BotRepository        port.BotRepository
	EventBus             port.EventBus
	InstanceManager      port.InstanceManager
	MessageSender        port.MessageSender
	RunRepository        port.RunRepository
	ScriptMetaProvider   port.ScriptMetaProvider
	ScriptRepository     port.ScriptRepository
	ThreadRepository     port.ThreadRepository
	ThreadsTableProvider port.ThreadsTableProvider
	UserRepository       port.UserRepository
}

func NewApplication(i Infra, l *slog.Logger) *Application {
	return &Application{
		Commands: Commands{
			CreateBot:    command.NewCreateBotHandler(i.BotRepository, i.ScriptMetaProvider, l),
			CreateRun:    command.NewCreateRunHandler(i.RunRepository, i.BotMetaProvider, i.EventBus, l),
			CreateScript: command.NewCreateScriptHandler(i.ScriptRepository, l),
			DeleteBot:    command.NewDeleteBotHandler(i.BotRepository, l),
			DeleteScript: command.NewDeleteScriptHandler(i.ScriptRepository, l),
			Entry: command.NewEntryHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.UserRepository, i.MessageSender, l,
			),
			Process: command.NewProcessHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.MessageSender, l,
			),
			UpdateBot:    command.NewUpdateBotHandler(i.BotRepository, i.ScriptMetaProvider, l),
			UpdateScript: command.NewUpdateScriptHandler(i.ScriptRepository, l),
		},
		Queries: Queries{
			GetBot:     query.NewGetBotHandler(i.BotRepository, l),
			GetBots:    query.NewGetBotsHandler(i.BotRepository, l),
			GetScript:  query.NewGetScriptHandler(i.ScriptRepository, l),
			GetScripts: query.NewGetScriptsHandler(i.ScriptRepository, l),
		},
		Events: EventHandlers{
			StartOnRunStartRequested: eventhandler.NewStartOnRunStartRequestedHandler(
				i.RunRepository, i.InstanceManager, i.EventBus, l,
			),
			StartOnRunRecoverRequested: eventhandler.NewStartOnRunRecoverRequestedHandler(
				i.RunRepository, i.InstanceManager, i.EventBus, l,
			),
		},
	}
}
