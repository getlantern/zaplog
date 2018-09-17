package zaplog

import (
	"fmt"
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

	reporters      []ErrorReporter
	reportersMutex sync.RWMutex

	reports = make(chan func(), 1000)
)

// ErrorReporter is a function to which the logger will report errors.
type ErrorReporter func(level zapcore.Level, msg string)

// SetZapConfig allows users to customize the configuration. Note this must be called before
// any logging takes place -- it will not reset the configuration of an existing logger.
func SetZapConfig(config zap.Config) {
	zapConfig = config
}

// LoggerFor creates a new zap logger with the specified name.
func LoggerFor(prefix string) *zap.SugaredLogger {
	once.Do(func() {
		baseLog, _ := zapConfig.Build()
		// Make sure our wrapper code isn't what always shows up as the caller.
		zlog = baseLog.Sugar()
	})
	return zlog.Named(prefix)
}

// Close closes the logger
func Close() {
	if zlog != nil {
		zlog.Sync()
	}
}

// RegisterReporter registers the given ErrorReporter. All logged Errors are
// sent to this reporter.
func RegisterReporter(reporter ErrorReporter) {
	reportersMutex.Lock()
	reporters = append(reporters, reporter)
	reportersMutex.Unlock()
}

func (wl *wrappedLogger) Warn(msg string, fields ...zap.Field) {
	wl.Warn(msg, fields...)
}

func (wl *wrappedLogger) Warnf(template string, args ...interface{}) {
	wl.Warnf(template, args...)
	wl.reportError(zap.WarnLevel, template, args...)
}

func (wl *wrappedLogger) Error(msg string, fields ...zap.Field) {
	wl.Error(msg, fields...)
}

func (wl *wrappedLogger) Errorf(template string, args ...interface{}) {
	wl.Errorf(template, args...)
	wl.reportError(zap.ErrorLevel, template, args...)
}

func (wl *wrappedLogger) Fatal(msg string, fields ...zap.Field) {
	wl.Fatal(msg, fields...)
}

func (wl *wrappedLogger) Fatalf(template string, args ...interface{}) {
	wl.Fatalf(template, args...)
	wl.reportError(zap.FatalLevel, template, args...)
}

func (wl *wrappedLogger) reportError(level zapcore.Level, template string, args ...interface{}) {
	reports <- func() {
		report(level, template, args...)
	}
}

func report(level zapcore.Level, template string, args ...interface{}) {
	var reportersCopy []ErrorReporter
	reportersMutex.RLock()
	if len(reporters) > 0 {
		reportersCopy = make([]ErrorReporter, len(reporters))
		copy(reportersCopy, reporters)
	}
	reportersMutex.RUnlock()

	msg := fmt.Sprintf(template, args...)
	for _, reporter := range reportersCopy {
		// We include globals when reporting
		reporter(level, msg)
	}
}
