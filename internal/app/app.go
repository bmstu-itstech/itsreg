package app

import (
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
)

type Commands struct {
	CreateBot    *command.CreateBotHandler
	CreateScript *command.CreateScriptHandler
	DeleteScript *command.DeleteScriptHandler
	Entry        *command.EntryHandler
	Process      *command.ProcessHandler
}

type Queries struct {
	GetBot     *query.GetBotHandler
	GetScript  *query.GetScriptHandler
	GetScripts *query.GetScriptsHandler
}

type Application struct {
	Commands Commands
	Queries  Queries
}

type Infra struct {
	BotRepository        port.BotRepository
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
			CreateScript: command.NewCreateScriptHandler(i.ScriptRepository, l),
			DeleteScript: command.NewDeleteScriptHandler(i.ScriptRepository, l),
			Entry: command.NewEntryHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.UserRepository, i.MessageSender, l,
			),
			Process: command.NewProcessHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.MessageSender, l,
			),
		},
		Queries: Queries{
			GetBot:     query.NewGetBotHandler(i.BotRepository, l),
			GetScript:  query.NewGetScriptHandler(i.ScriptRepository, l),
			GetScripts: query.NewGetScriptsHandler(i.ScriptRepository, l),
		},
	}
}
