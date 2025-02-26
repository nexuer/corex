package log

import (
	"context"
	"encoding"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

type TextHandler struct {
	handler *commonHandler
}

func Text() Handler {
	return &TextHandler{
		handler: newCommonHandler("", false),
	}
}

func TextWithName(name string) Handler {
	return &TextHandler{
		handler: newCommonHandler(name, false),
	}
}

func (h *TextHandler) With(kvs ...any) Handler {
	return &TextHandler{
		handler: h.handler.withAttrs(kvsToAttrs(kvs)),
	}
}

func (h *TextHandler) Log(ctx context.Context, w io.Writer, level Level, msg string, kvs ...any) error {
	return h.handler.handle(ctx, w, level, msg, kvs...)
}

// byteSlice returns its argument as a []byte if the argument's
// underlying type is []byte, along with a second return value of true.
// Otherwise it returns nil, false.
func byteSlice(a any) ([]byte, bool) {
	if bs, ok := a.([]byte); ok {
		return bs, true
	}
	// Like Printf's %s, we allow both the slice type and the byte element type to be named.
	t := reflect.TypeOf(a)
	if t != nil && t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return reflect.ValueOf(a).Bytes(), true
	}
	return nil, false
}

func bytesToString(data []byte) string {
	return unsafe.String(unsafe.SliceData(data), len(data))
}

func appendTextValue(s *handleState, v Value) error {
	switch v.Kind() {
	case KindString:
		s.appendString(v.String())
	case KindTime:
		s.appendTime(v.Time())
	case KindInt64:
		*s.buf = strconv.AppendInt(*s.buf, v.Int64(), 10)
	case KindUint64:
		*s.buf = strconv.AppendUint(*s.buf, v.Uint64(), 10)
	case KindFloat64:
		*s.buf = strconv.AppendFloat(*s.buf, v.Float64(), 'g', -1, 64)
	case KindBool:
		*s.buf = strconv.AppendBool(*s.buf, v.Bool())
	case KindDuration:
		*s.buf = strconv.AppendInt(*s.buf, v.Duration().Nanoseconds(), 10)
	case KindAny:
		switch anyVal := v.Any().(type) {
		case error:
			s.appendString(anyVal.Error())
		case encoding.TextMarshaler:
			data, err := anyVal.MarshalText()
			if err != nil {
				return err
			}
			// TODO: avoid the conversion to string.
			s.appendString(bytesToString(data))
		default:
			s.appendString(Sprintf("%+v", anyVal))
		}
	}
	return nil
}
