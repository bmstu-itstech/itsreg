package app

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/app/command"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/query"
)

type Commands struct {
	CreateBot    command.CreateBotHandler
	DisableBot   command.DisableBotHandler
	EnableBot    command.EnableBotHandler
	Entry        command.EntryHandler
	Mailing      command.MailingHandler
	Process      command.ProcessHandler
	Start        command.StartHandler
	StartEnabled command.StartEnabledHandler
	Stop         command.StopHandler
	UpdateBot    command.UpdateBotHandler
}

type Queries struct {
	GetBot      query.GetBotHandler
	GetStatus   query.GetStatusHandler
	GetThreads  query.GetThreadsHandler
	GetUserBots query.GetUserBotsHandler
}

type Application struct {
	Commands Commands
	Queries  Queries
}
