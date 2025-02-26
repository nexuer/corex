package log

import (
	"context"
	"errors"
	"io"

	klog "github.com/go-kratos/kratos/v2/log"
)

var discard = writerWrapper{Writer: io.Discard}

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Close() error {
	return nil
}

type Handler interface {
	With(kvs ...any) Handler
	Log(ctx context.Context, w io.Writer, level Level, msg string, kvs ...any) error
}

type Logger struct {
	ctx     context.Context
	w       io.WriteCloser
	level   Level
	handler Handler
}

// AddSync converts an io.Writer to a WriteSyncer. It attempts to be
// intelligent: if the concrete type of the io.Writer implements WriteSyncer,
// we'll use the existing Sync method. If it doesn't, we'll add a no-op Sync.
func addWrite(w io.Writer) io.WriteCloser {
	if w == nil {
		return nil
	}
	switch w := w.(type) {
	case io.WriteCloser:
		return w
	default:
		return writerWrapper{w}
	}
}

func New(w io.Writer, h ...Handler) *Logger {
	l := &Logger{
		ctx: context.Background(),
		w:   addWrite(w),
	}
	if len(h) == 0 {
		l.handler = Text()
	} else {
		l.handler = h[0]
	}
	return l
}

func (l *Logger) clone() *Logger {
	c := &Logger{
		ctx:     l.ctx,
		w:       l.w,
		level:   l.level,
		handler: l.handler,
	}
	return c
}

func (l *Logger) Close() error {
	if l.w == nil {
		return nil
	}
	return l.w.Close()
}

func (l *Logger) SetLevel(level Level) {
	if level == l.level {
		return
	}
	l.level = level
}

func (l *Logger) SetOutput(w io.Writer) {
	if l.Writer() == w {
		return
	}
	l.w = addWrite(w)
}

func (l *Logger) SetHandler(h Handler) {
	if l.handler == h {
		return
	}
	l.handler = h
}

func (l *Logger) enable(level Level) bool {
	return level >= l.level
}

func (l *Logger) log(level Level, msg string, kvs ...any) error {
	if msg == "" && len(kvs) == 0 {
		return nil
	}

	if !l.enable(level) {
		return nil
	}

	if l.handler == nil {
		return errors.New("logger handler is nil")
	}

	if err := l.handler.Log(l.ctx, l.w, level, msg, kvs...); err != nil {
		return err
	}

	if afterWrite != nil {
		afterWrite(level)
	}

	return nil
}

var (
	kLevelMap = map[klog.Level]Level{
		klog.LevelDebug: LevelDebug,
		klog.LevelInfo:  LevelInfo,
		klog.LevelWarn:  LevelWarn,
		klog.LevelError: LevelError,
		klog.LevelFatal: LevelFatal,
	}
)

func (l *Logger) Log(level klog.Level, kvs ...any) error {
	var msg string
	if len(kvs) >= 2 {
		msg, _ = kvs[1].(string)
		kvs = kvs[2:]
		return l.log(kLevelMap[level], msg, kvs...)
	}
	return l.log(kLevelMap[level], "")
}

func (l *Logger) With(kvs ...any) *Logger {
	if len(kvs) == 0 || l.handler == nil {
		return l
	}
	l2 := l.clone()
	l2.handler = l.handler.With(kvs...)
	return l2
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	l2 := l.clone()
	l2.ctx = ctx
	return l2
}

func (l *Logger) Writer() io.Writer {
	switch w := l.w.(type) {
	case writerWrapper:
		return w.Writer
	default:
		return l.w
	}
}

// Debug logs a message at debug level.
func (l *Logger) Debug(args ...any) {
	err := l.log(LevelDebug, getMessage("", args))
	errorHandler(err)
}

// Debugf logs a message at debug level.
func (l *Logger) Debugf(format string, args ...any) {
	err := l.log(LevelDebug, getMessage(format, args))
	errorHandler(err)
}

// DebugS logs a message at debug level with kvs.
func (l *Logger) DebugS(msg string, kvs ...any) {
	err := l.log(LevelDebug, msg, kvs...)
	errorHandler(err)
}

// Info logs a message at info level.
func (l *Logger) Info(args ...any) {
	err := l.log(LevelInfo, getMessage("", args))
	errorHandler(err)
}

// Infof logs a message at info level.
func (l *Logger) Infof(format string, args ...any) {
	err := l.log(LevelInfo, getMessage(format, args))
	errorHandler(err)
}

// InfoS logs a message at info level with kvs.
func (l *Logger) InfoS(msg string, kvs ...any) {
	err := l.log(LevelInfo, msg, kvs...)
	errorHandler(err)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(args ...any) {
	err := l.log(LevelWarn, getMessage("", args))
	errorHandler(err)
}

// Warnf logs a message at warn level.
func (l *Logger) Warnf(format string, args ...any) {
	err := l.log(LevelWarn, getMessage(format, args))
	errorHandler(err)
}

// WarnS logs a message at warn level with kvs.
func (l *Logger) WarnS(msg string, kvs ...any) {
	err := l.log(LevelWarn, msg, kvs...)
	errorHandler(err)
}

// Error logs a message at error level.
func (l *Logger) Error(args ...any) {
	err := l.log(LevelError, getMessage("", args))
	errorHandler(err)
}

// Errorf logs a message at warn level.
func (l *Logger) Errorf(format string, args ...any) {
	err := l.log(LevelError, getMessage(format, args))
	errorHandler(err)
}

// ErrorS logs a message at error level with kvs.
func (l *Logger) ErrorS(err error, msg string, kvs ...any) {
	if err == nil {
		errorHandler(l.log(LevelError, msg, kvs...))
		return
	}
	nv := make([]any, 0, len(kvs)+2)
	nv = append(nv, ErrKey, err.Error())
	nv = append(nv, kvs...)
	errorHandler(l.log(LevelError, msg, nv...))
}

// Fatal logs a message at fatal level.
func (l *Logger) Fatal(args ...any) {
	err := l.log(LevelFatal, getMessage("", args))
	errorHandler(err)
}

// Fatalf logs a message at warn level.
func (l *Logger) Fatalf(format string, args ...any) {
	err := l.log(LevelFatal, getMessage(format, args))
	errorHandler(err)
}

// FatalS logs a message at fatal level with kvs.
func (l *Logger) FatalS(msg string, kvs ...any) {
	err := l.log(LevelFatal, msg, kvs...)
	errorHandler(err)
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, args []any) string {
	if len(args) == 0 {
		return template
	}

	if template != "" {
		return Sprintf(template, args...)
	}

	if len(args) == 1 {
		if str, ok := args[0].(string); ok {
			return str
		}
	}

	return Sprint(args...)
}
