package log

import (
	"log/slog"
	"testing"
)

func BenchmarkJsonInfo(b *testing.B) {
	l := New(output, Json())
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info(fakeMessage)
		}
	})
}

func BenchmarkJsonInfoWith(b *testing.B) {
	l := New(output, Json()).With(
		"ts", DefaultTimestamp,
		"caller", DefaultCaller,
		"key1", "value1",
		"key2", "value2",
		"key3", "value3",
		"key4", "value4",
		"key5", "value5",
		"key6", "value6",
		"key7", "value7",
		"key8", "value8",
		"key9", "value9",
		"key10", "value10",
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			l.InfoS(fakeMessage,
				"key11", "value11",
				"key12", "value12",
				"key13", "value13",
				"key14", "value14",
				"key15", "value15",
				"key16", "value16",
				"key17", "value17",
				"key18", "value18",
				"key19", "value19",
				"key20", "value20",
			)
		}
	})
}

func BenchmarkSlogJsonWith(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
		//ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		//	if a.Key == slog.LevelKey && a.Value.String() == "ERROR+119" {
		//		return slog.String(a.Key, "FATAL")
		//	}
		//	return a
		//},
	})).With(
		"ts", DefaultTimestamp,
		"caller", DefaultCaller,
		"key1", "value1",
		"key2", "value2",
		"key3", "value3",
		"key4", "value4",
		"key5", "value5",
		"key6", "value6",
		"key7", "value7",
		"key8", "value8",
		"key9", "value9",
		"key10", "value10",
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info(fakeMessage,
				"key11", "value11",
				"key12", "value12",
				"key13", "value13",
				"key14", "value14",
				"key15", "value15",
				"key16", "value16",
				"key17", "value17",
				"key18", "value18",
				"key19", "value19",
				"key20", "value20",
			)
		}
	})
}
