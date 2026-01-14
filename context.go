package grepo

import (
	"context"
	"time"
)

type ctxkey string

const (
	ctxkeyExecuteTime ctxkey = "ExecuteTime"
)

func ExecuteTime(ctx context.Context) time.Time {
	if v := ctx.Value(ctxkeyExecuteTime); v != nil {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

func WithExecuteTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxkeyExecuteTime, t)
}
