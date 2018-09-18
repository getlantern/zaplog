package zaplog

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type wrappedLogger struct {
	*zap.SugaredLogger
}

var (
	zapConfig = zap.NewProductionConfig()
	zlog      *zap.SugaredLogger
	once      sync.Once

	hookFuncs  []Hook
	hooksMutex sync.RWMutex

	hooks = make(chan func(), 1000)
)

// Hook is a function to which the logger will report logs.
type Hook func(entry zapcore.Entry) error

// SetZapConfig allows users to customize the configuration. Note this must be called before
// any logging takes place -- it will not reset the configuration of an existing logger.
func SetZapConfig(config zap.Config) {
	zapConfig = config
}

// LoggerFor creates a new zap logger with the specified name.
func LoggerFor(prefix string) *zap.SugaredLogger {
	once.Do(func() {
		baseLog, _ := zapConfig.Build()
		baseLog = baseLog.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
			if entry.Level < zapcore.WarnLevel {
				return nil
			}
			hook(entry)
			return nil
		}))
		// Make sure our wrapper code isn't what always shows up as the caller.
		zlog = baseLog.Sugar()
		go processHooks()
	})
	return zlog.Named(prefix)
}

func processHooks() {
	for hook := range hooks {
		hook()
	}
}

// Close closes the logger
func Close() {
	if zlog != nil {
		zlog.Sync()
	}
	close(hooks)
}

// AddWarnHook registers the given Hook that will be called for warn level logs or above.
func AddWarnHook(h Hook) {
	hooksMutex.Lock()
	hookFuncs = append(hookFuncs, h)
	hooksMutex.Unlock()
}

func hook(entry zapcore.Entry) {
	hooks <- func() {
		hooksMutex.Lock()
		for _, hookFunc := range hookFuncs {
			hookFunc(entry)
		}
		hooksMutex.Unlock()
	}
}
