package hooks

import (
	"context"
	"log/slog"

	"github.com/ralsnet/grepo"
)

type HookSlogOptions struct {
	level slog.Level
	msg   string
}

type HookSlogOptionFunc func(*HookSlogOptions)

func WithSlogLevel(level slog.Level) HookSlogOptionFunc {
	return func(o *HookSlogOptions) {
		o.level = level
	}
}

func WithSlogMsg(msg string) HookSlogOptionFunc {
	return func(o *HookSlogOptions) {
		o.msg = msg
	}
}

func HookBeforeSlog(opts ...HookSlogOptionFunc) grepo.BeforeHook[any] {
	options := &HookSlogOptions{
		level: slog.LevelInfo,
		msg:   "Starting operation",
	}
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, desc grepo.Descriptor, i any) (context.Context, error) {
		slog.Log(ctx, options.level, options.msg, "operation", desc.Operation(), "input", i)
		return ctx, nil
	}
}

func HookAfterSlog(opts ...HookSlogOptionFunc) grepo.AfterHook[any, any] {
	options := &HookSlogOptions{
		level: slog.LevelInfo,
		msg:   "Finished operation",
	}
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, desc grepo.Descriptor, i any, o any) {
		slog.Log(ctx, options.level, options.msg, "operation", desc.Operation(), "output", o)
	}
}

func HookErrorSlog(opts ...HookSlogOptionFunc) grepo.ErrorHook[any] {
	options := &HookSlogOptions{
		level: slog.LevelError,
		msg:   "Operation error",
	}
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, desc grepo.Descriptor, i any, e error) {
		slog.Log(ctx, options.level, options.msg, "operation", desc.Operation(), "input", i, "error", e)
	}
}
