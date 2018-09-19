package zaplog

import (
	"os"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type wrappedLogger struct {
	*zap.SugaredLogger
}

var (
	zapConfig atomic.Value
	zlog      *zap.SugaredLogger
	once      sync.Once

	hookFuncs  []Hook
	hooksMutex sync.RWMutex

	hooks   = make(chan func(), 1000)
	onFatal atomic.Value
)

// Hook is a function to which the logger will report logs.
type Hook func(entry zapcore.Entry) error

// SetZapConfig allows users to customize the configuration. Note this must be called before
// any logging takes place -- it will not reset the configuration of an existing logger.
func SetZapConfig(config zap.Config) {
	zapConfig.Store(config)
}

// LoggerFor creates a new zap logger with the specified name.
func LoggerFor(prefix string) *zap.SugaredLogger {
	once.Do(func() {
		conf := zapConfig.Load()
		if conf == nil {
			zapConfig.Store(zap.NewProductionConfig())
		}
		baseLog, _ := zapConfig.Load().(zap.Config).Build()
		baseLog = baseLog.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
			if entry.Level < zapcore.WarnLevel {
				return nil
			}
			if entry.Level == zapcore.FatalLevel {
				fn := onFatal.Load().(func())
				fn()
			} else {
				hook(entry)
			}
			return nil
		}))
		onFatal.Store(func() {
			os.Exit(1)
		})
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

// OnFatal configures golog to call the given function on any FATAL error. By
// default, golog calls os.Exit(1) on any FATAL error.
func OnFatal(fn func()) {
	onFatal.Store(fn)
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
