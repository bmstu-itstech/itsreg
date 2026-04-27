package reqctx

import "context"

type ctxKey int

const ctxKeyRequestID ctxKey = iota

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, requestID)
}

func FromContext(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(ctxKeyRequestID).(string)
	return reqID, ok
}
