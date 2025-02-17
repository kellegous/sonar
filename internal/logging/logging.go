package logging

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey = contextKey("logger")

func L(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
			return l
		}
	}
	return zap.L()
}

func With(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}
