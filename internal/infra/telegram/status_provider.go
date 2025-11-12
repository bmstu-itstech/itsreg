package telegram

import (
	"context"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func (m *InstanceManager) Status(_ context.Context, id bots.BotID) (bots.Status, error) {
	r, ok := m.m.Load(id)
	if !ok {
		return bots.Idle, nil
	}

	ins, _ := r.(*botInstance)
	if ins.IsDead() {
		return bots.Dead, nil
	}
	return bots.Running, nil
}
