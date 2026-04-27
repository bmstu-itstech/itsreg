package app

import (
	"log/slog"

	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/eventhandler"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
)

type Commands struct {
	CreateBot     *command.CreateBotHandler
	CreateMailing *command.CreateMailingHandler
	CreateRun     *command.CreateRunHandler
	CreateScript  *command.CreateScriptHandler
	DeleteBot     *command.DeleteBotHandler
	DeleteScript  *command.DeleteScriptHandler
	Entry         *command.EntryHandler
	Process       *command.ProcessHandler
	StopRun       *command.StopRunHandler
	UpdateBot     *command.UpdateBotHandler
	UpdateScript  *command.UpdateScriptHandler
}

type Queries struct {
	GetBot         *query.GetBotHandler
	GetBotMailings *query.GetBotMailingsHandler
	GetBotRuns     *query.GetBotRunsHandler
	GetBots        *query.GetBotsHandler
	GetMailing     *query.GetMailingHandler
	GetMailings    *query.GetMailingsHandler
	GetRun         *query.GetRunHandler
	GetRuns        *query.GetRunsHandler
	GetScript      *query.GetScriptHandler
	GetScripts     *query.GetScriptsHandler
}

type EventHandlers struct {
	SendMailingMessage         *eventhandler.SendMailingMessageHandler
	SendOnSendMessageRequested *eventhandler.SendOnSendMessageRequestedHandler
	StartOnRunStartRequested   *eventhandler.StartOnRunStartRequestedHandler
	StartOnRunRecoverRequested *eventhandler.StartOnRunRecoverRequestedHandler
	StartScheduledMailing      *eventhandler.StartScheduledMailingHandler
	StopOnRunStopRequested     *eventhandler.StopOnRunStopRequestedHandler
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
	MailingRepository    port.MailingRepository
	MessageSender        port.MessageSender
	OwnedMailingProvider port.OwnedMailingProvider
	OwnedRunProvider     port.OwnedRunProvider
	RateLimiter          port.RateLimiter
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
			CreateBot:     command.NewCreateBotHandler(i.BotRepository, i.ScriptMetaProvider, l),
			CreateMailing: command.NewCreateMailingHandler(i.MailingRepository, i.BotMetaProvider, i.EventBus, l),
			CreateRun:     command.NewCreateRunHandler(i.RunRepository, i.BotMetaProvider, i.EventBus, l),
			CreateScript:  command.NewCreateScriptHandler(i.ScriptRepository, l),
			DeleteBot:     command.NewDeleteBotHandler(i.BotRepository, l),
			DeleteScript:  command.NewDeleteScriptHandler(i.ScriptRepository, l),
			Entry: command.NewEntryHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.UserRepository, i.EventBus, l,
			),
			Process: command.NewProcessHandler(
				i.BotRepository, i.ScriptRepository, i.ThreadRepository, i.EventBus, l,
			),
			StopRun:      command.NewStopRunHandler(i.RunRepository, i.BotMetaProvider, i.EventBus, l),
			UpdateBot:    command.NewUpdateBotHandler(i.BotRepository, i.ScriptMetaProvider, l),
			UpdateScript: command.NewUpdateScriptHandler(i.ScriptRepository, l),
		},
		Queries: Queries{
			GetBot:         query.NewGetBotHandler(i.BotRepository, l),
			GetBotMailings: query.NewGetBotMailingsHandler(i.MailingRepository, i.BotMetaProvider, l),
			GetBotRuns:     query.NewGetBotRunsHandler(i.RunRepository, i.BotMetaProvider, l),
			GetBots:        query.NewGetBotsHandler(i.BotRepository, l),
			GetMailing:     query.NewGetMailingHandler(i.OwnedMailingProvider, l),
			GetMailings:    query.NewGetMailingsHandler(i.MailingRepository, l),
			GetRun:         query.NewGetRunHandler(i.OwnedRunProvider, l),
			GetRuns:        query.NewGetRunsHandler(i.RunRepository, l),
			GetScript:      query.NewGetScriptHandler(i.ScriptRepository, l),
			GetScripts:     query.NewGetScriptsHandler(i.ScriptRepository, l),
		},
		Events: EventHandlers{
			SendMailingMessage: eventhandler.NewSendMailingMessageHandler(
				i.MessageSender, i.MailingRepository, i.BotMetaProvider, i.RateLimiter, i.EventBus, l,
			),
			SendOnSendMessageRequested: eventhandler.NewSendOnSendMessageRequestedHandler(
				i.MessageSender, i.BotMetaProvider, i.RateLimiter, i.EventBus, l,
			),
			StartOnRunStartRequested: eventhandler.NewStartOnRunStartRequestedHandler(
				i.RunRepository, i.InstanceManager, i.EventBus, l,
			),
			StartOnRunRecoverRequested: eventhandler.NewStartOnRunRecoverRequestedHandler(
				i.RunRepository, i.InstanceManager, i.EventBus, l,
			),
			StartScheduledMailing: eventhandler.NewStartScheduledMailing(
				i.MailingRepository, i.ThreadRepository, i.BotMetaProvider, i.ScriptRepository, i.EventBus, l,
			),
			StopOnRunStopRequested: eventhandler.NewStopOnRunStopRequestedHandler(
				i.RunRepository, i.InstanceManager, i.EventBus, l,
			),
		},
	}
}
