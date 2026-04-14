package jwtauth

import (
	"context"
)

type ctxKey int

const ctxKeyUID ctxKey = iota

func toContext(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, ctxKeyUID, uid)
}

func FromContext(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(ctxKeyUID).(int64)
	return uid, ok
}
