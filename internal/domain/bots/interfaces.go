package bots

import (
	"context"
	"errors"
)

type BotManager interface {
	// Upsert создаёт нового бота или обновляет существующий с данным BotID.
	Upsert(ctx context.Context, bot Bot) error
}

var ErrBotNotFound = errors.New("bot not found")

type BotProvider interface {
	// Bot возвращает найденного бота или ошибку ErrBotNotFound.
	Bot(ctx context.Context, id BotID) (Bot, error)

	// UserBots возвращает возможно пустой список ботов, чьи авторы имеют UserID.
	UserBots(ctx context.Context, userID UserID) ([]Bot, error)
}

type ParticipantRepository interface {
	// UpdateOrCreate обновляет Participant через callback-функцию updateFn.
	// Создаёт Participant с заданным ParticipantID, если таковой не существует.
	UpdateOrCreate(
		ctx context.Context,
		id ParticipantID,
		updateFn func(context.Context, *Participant) error,
	) error
}

type Username string

type UsernameManager interface {
	// Upsert добавляет или обновляет username для заданного UserID.
	Upsert(ctx context.Context, id UserID, username Username) error
}

var ErrUsernameNotFound = errors.New("username not found")

type UsernameProvider interface {
	// Username возвращает имя пользователя с UserID.
	Username(ctx context.Context, id UserID) (Username, error)
}

type ThreadProvider interface {
	// BotThreads возвращает все цепочки ответов (треды) заданному боту.
	BotThreads(ctx context.Context, botID BotID) ([]BotThread, error)
}

type BotMessageSender interface {
	Send(ctx context.Context, token Token, userID UserID, msg BotMessage) error
}

var ErrRunningInstanceNotFound = errors.New("running instance not found")

type InstanceManager interface {
	Start(ctx context.Context, id BotID, token Token) error
	Stop(ctx context.Context, id BotID) error
}

type StatusManager interface {
	UpdateStatus(ctx context.Context, id BotID, status Status) error
}

type StatusProvider interface {
	Status(ctx context.Context, id BotID) (Status, error)
}

type ProcessHandler interface {
	Process(ctx context.Context, botID BotID, userID UserID, msg Message) error
}

type EntryHandler interface {
	Entry(ctx context.Context, botID BotID, userID UserID, key EntryKey) error
}
