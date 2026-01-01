package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

const (
	loggerKey = "logger"
)

type Logger struct {
	*zap.Logger
}

func New() (*Logger, error) {
	lg, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	return &Logger{
		lg,
	}, nil
}

func (l *Logger) Sync() {
	l.Logger.Sync()
}

func WithContext(ctx context.Context, lg *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, lg)
}

func FromContext(ctx context.Context) (*Logger, bool) {
	lg, ok := ctx.Value(loggerKey).(*Logger)
	return lg, ok
}
