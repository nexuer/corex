package log

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Value = slog.Value

var (
	// DefaultCaller is a Valuer that returns the file and line.
	DefaultCaller = Caller(6)

	// DefaultTimestamp is a Valuer that returns the current wallclock time.
	DefaultTimestamp = Timestamp(time.RFC3339)
)

type Valuer func(ctx context.Context) Value

func Timestamp(layout string) Valuer {
	return func(ctx context.Context) Value {
		return slog.StringValue(time.Now().Format(layout))
	}
}

// ValuerValue return the function value.
func ValuerValue(ctx context.Context, v Valuer) Value {
	return v(ctx)
}

func Caller(depth int) Valuer {
	return func(ctx context.Context) Value {
		_, file, line, _ := runtime.Caller(depth)
		idx := strings.LastIndexByte(file, '/')
		if idx == -1 {
			return slog.StringValue(file[idx+1:] + ":" + strconv.Itoa(line))
		}
		idx = strings.LastIndexByte(file[:idx], '/')
		return slog.StringValue(file[idx+1:] + ":" + strconv.Itoa(line))
	}
}
