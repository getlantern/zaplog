package zaplog

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zlog     *zap.SugaredLogger
	logMutex sync.RWMutex
)

// Hook is a function to which the logger will report logs.
type Hook func(entry zapcore.Entry) error

func init() {
	baseLog, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zlog = baseLog.Sugar()
}

// LoggerFor creates a new zap logger with the specified name.
func LoggerFor(prefix string, config ...zap.Config) *zap.SugaredLogger {
	return zlog.Named(prefix)
}

func SetConfig(config zap.Config) error {
	logMutex.Lock()
	defer logMutex.Unlock()
	baseLog, err := config.Build()
	if err != nil {
		return fmt.Errorf("could not build logger from config %w", err)
	}
	zlog = baseLog.Sugar()
	return nil
}

// RegisterHook registers the given Hook for listening to log events.
func RegisterHook(h Hook) {
	logMutex.Lock()
	defer logMutex.Unlock()
	baseLog := zlog.Desugar().WithOptions(zap.Hooks(h))
	zlog = baseLog.Sugar()
}
