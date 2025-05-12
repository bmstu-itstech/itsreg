package app

type Commands struct {
	CreateBot      CreateBotHandler
	Entry          EntryHandler
	Process        ProcessHandler
	Start          StartHandler
	Stop           StopHandler
	UpdateBot      UpdateBotHandler
	UpdateUsername UpdateUsernameHandler
}

type Queries struct {
	GetBot      GetBotHandler
	GetStatus   GetStatusHandler
	GetThreads  GetThreadsHandler
	GetUserBots GetUserBotsHandler
	GetUsername GetUsernameHandler
}

type Application struct {
	Commands Commands
	Queries  Queries
}
