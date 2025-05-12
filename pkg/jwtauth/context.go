package jwtauth

import (
	"context"
	"errors"
)

type ctxKey int

const userCtxKey ctxKey = iota

func userUUIDToContext(ctx context.Context, userUUID string) context.Context {
	return context.WithValue(ctx, userCtxKey, userUUID)
}

var ErrNoUserUUIDInContext = errors.New("no user uuid in context")

func UserUUIDFromContext(ctx context.Context) (string, error) {
	u, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		return "", ErrNoUserUUIDInContext
	}
	return u, nil
}
