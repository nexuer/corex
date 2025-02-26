package log

import (
	"log/slog"
)

type attr = slog.Attr

type Kind = slog.Kind

const badKey = "<BAD_KEY>"

const (
	KindAny       = slog.KindAny
	KindBool      = slog.KindBool
	KindDuration  = slog.KindDuration
	KindFloat64   = slog.KindFloat64
	KindInt64     = slog.KindInt64
	KindString    = slog.KindString
	KindTime      = slog.KindTime
	KindUint64    = slog.KindUint64
	KindGroup     = slog.KindGroup
	KindLogValuer = slog.KindLogValuer
)

func isEmpty(a attr) (t bool) {
	if a.Key == "" {
		return true
	}
	// empty group
	if a.Value.Kind() == KindGroup && len(a.Value.Group()) == 0 {
		return true
	}
	return false
}

var (
	Group = slog.Group
	//String   = slog.String
	//Int      = slog.Int
	//Int64    = slog.Int64
	//Uint64   = slog.Uint64
	//Float64  = slog.Float64
	//Bool     = slog.Bool
	//Time     = slog.Time
	//Duration = slog.Duration
	//Any      = slog.Any
)

func kvsToAttrs(kvs []any) []attr {
	var (
		a     attr
		attrs []attr
	)

	for len(kvs) > 0 {
		a, kvs = kvsToAttr(kvs)
		attrs = append(attrs, a)
	}
	return attrs
}

// kvsToAttr turns a prefix of the nonempty args slice into an Attr
// and returns the unconsumed portion of the slice.
// If args[0] is an Attr, it returns it.
// If args[0] is a string, it treats the first two elements as
// a key-value pair.
// Otherwise, it treats args[0] as a value with a missing key.
func kvsToAttr(args []any) (attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case attr:
		return x, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}
