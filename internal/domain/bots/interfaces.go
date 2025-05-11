package bots

import (
	"context"
	"errors"
)

type BotManager interface {
	// Upsert создаёт нового бота или обновляет существующий с данным BotId.
	Upsert(ctx context.Context, bot Bot) error
}

var ErrBotNotFound = errors.New("bot not found")

type BotProvider interface {
	// Bot возвращает найденного бота или ошибку ErrBotNotFound.
	Bot(ctx context.Context, id BotId) (Bot, error)

	// UserBots возвращает возможно пустой список ботов, чьи авторы имеют userId.
	UserBots(ctx context.Context, userId UserId) ([]Bot, error)
}

type ParticipantRepository interface {
	// UpdateOrCreate обновляет Participant через callback-функцию updateFn.
	// Создаёт Participant с заданным ParticipantId, если таковой не существует.
	UpdateOrCreate(
		ctx context.Context,
		id ParticipantId,
		updateFn func(context.Context, *Participant) error,
	) error
}

type Username string

type UsernameManager interface {
	// Upsert добавляет или обновляет username для заданного UserId.
	Upsert(ctx context.Context, id UserId, username Username) error
}

var ErrUsernameNotFound = errors.New("username not found")

type UsernameProvider interface {
	// Username возвращает имя пользователя с UserId.
	Username(ctx context.Context, id UserId) (Username, error)
}

type ThreadProvider interface {
	// BotThreads возвращает все цепочки ответов (треды) заданному боту.
	BotThreads(ctx context.Context, botId BotId) ([]Thread, error)
}

type BotMessageSender interface {
	Send(ctx context.Context, token Token, userId UserId, msg BotMessage) error
}
