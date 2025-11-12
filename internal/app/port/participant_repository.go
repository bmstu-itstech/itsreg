package port

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type ParticipantRepository interface {
	// UpdateOrCreateParticipant обновляет Participant через callback-функцию updateFn.
	// Создаёт Participant с заданным ParticipantID, если таковой не существует.
	UpdateOrCreateParticipant(
		ctx context.Context,
		id bots.ParticipantID,
		updateFn func(context.Context, *bots.Participant) error,
	) error
}
