package log

import (
	"fmt"
	"io"
	"os"

	klog "github.com/go-kratos/kratos/v2/log"
)

type hook func(level Level)

var (
	MessageKey = "msg"
	LevelKey   = "level"
	ErrKey     = "err"
	NameKey    = "logger"

	// ErrorHandler is called whenever fails to write an event on its
	// output. default an error is printed on the stderr. This handler must
	// be thread safe and non-blocking.
	ErrorHandler func(err error) = func(err error) {
		_, _ = fmt.Fprintf(os.Stderr, "log: write failed, %v\n", err)
	}

	afterWrite hook = func(level Level) {
		switch level {
		case LevelFatal:
			os.Exit(1)
		}
	}

	// Sprint is used to concatenate log message components into a single string.
	// It formats its arguments using the default behavior (fmt.Sprint).
	// The caller can replace this function to customize the log message formatting.
	Sprint func(a ...any) string = fmt.Sprint
	// Sprintf is used to format log messages with a specific format string.
	// It formats its arguments using the default behavior (fmt.Sprintf).
	// The caller can replace this function to customize the format or behavior of the log messages.
	Sprintf func(format string, a ...any) string = fmt.Sprintf
)

func errorHandler(err error) {
	if err == nil {
		return
	}
	if ErrorHandler != nil {
		ErrorHandler(err)
	}
}

var globalLogger *Logger

func init() {
	logger := New(os.Stderr).With("ts", DefaultTimestamp, "caller", DefaultCaller)
	SetLogger(logger)
}

func SetLogger(l *Logger) {
	if l == nil {
		return
	}
	globalLogger = l
	klog.SetLogger(globalLogger)
}

func GetLogger() *Logger {
	return globalLogger
}

func Close() error {
	if globalManager != nil {
		return globalManager.Close()
	}
	return globalLogger.Close()
}

func SetLevel(l Level) {
	globalLogger.SetLevel(l)
}

func SetOutput(w io.Writer) {
	globalLogger.SetOutput(w)
}

func SetHandler(h Handler) {
	globalLogger.SetHandler(h)
}

func Writer() io.Writer {
	return globalLogger.Writer()
}

// Debug logs a message at debug level.
func Debug(args ...any) {
	globalLogger.Debug(args...)
}

// Debugf logs a message at debug level.
func Debugf(format string, args ...any) {
	globalLogger.Debugf(format, args...)
}

// DebugS logs a message at debug level with key vals.
func DebugS(msg string, kvs ...any) {
	globalLogger.DebugS(msg, kvs...)
}

// Info logs a message at info level.
func Info(args ...any) {
	globalLogger.Info(args...)
}

// Infof logs a message at info level.
func Infof(format string, args ...any) {
	globalLogger.Infof(format, args...)
}

// InfoS logs a message at info level with key vals.
func InfoS(msg string, kvs ...any) {
	globalLogger.InfoS(msg, kvs...)
}

// Warn logs a message at warn level.
func Warn(args ...any) {
	globalLogger.Warn(args...)
}

// Warnf logs a message at warn level.
func Warnf(format string, args ...any) {
	globalLogger.Warnf(format, args...)
}

// WarnS logs a message at warn level with key vals.
func WarnS(msg string, kvs ...any) {
	globalLogger.WarnS(msg, kvs...)
}

// Error logs a message at error level.
func Error(args ...any) {
	globalLogger.Error(args...)
}

// Errorf logs a message at error level.
func Errorf(format string, args ...any) {
	globalLogger.Errorf(format, args...)
}

// ErrorS logs a message at error level with key vals.
func ErrorS(err error, msg string, kvs ...any) {
	globalLogger.ErrorS(err, msg, kvs...)
}

// Fatal logs a message at fatal level.
func Fatal(args ...any) {
	globalLogger.Fatal(args...)
}

// Fatalf logs a message at fatal level.
func Fatalf(format string, args ...any) {
	globalLogger.Fatalf(format, args...)
}

// FatalS logs a message at fatal level with key vals.
func FatalS(msg string, kvs ...any) {
	globalLogger.FatalS(msg, kvs...)
}

// log manager
var globalManager *Manager

// InitManager init global manager
func InitManager(name string, kvs ...any) {
	globalManager = NewManager(name, kvs...)
	SetLogger(globalManager.Get())
}

func M() *Manager {
	if globalManager == nil {
		panic("log: uninitialized manager not (forgotten use log.InitManager(name)?")
	}
	return globalManager
}

func L(name ...string) *Logger {
	return M().Get(name...)
}
