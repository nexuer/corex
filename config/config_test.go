package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	kratosconfig "github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type testConfig struct {
	Server struct {
		Addr string `json:"addr"`
	} `json:"server"`
	Enabled bool   `json:"enabled"`
	Count   int    `json:"count"`
	Name    string `json:"name"`
	Source  string `json:"source"`
}

func writeConfig(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func loadConfig(t *testing.T, dst any, opts ...Option) {
	t.Helper()
	closer, err := Load(dst, opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := closer.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestLoadPriority(t *testing.T) {
	cfg := testConfig{Enabled: true, Count: 10, Name: "default", Source: "default-source"}
	cfg.Server.Addr = "default"
	first := writeConfig(t, "first.yaml", "server: {addr: first}\nenabled: false\ncount: 0\nname: \"\"\n")
	second := writeConfig(t, "second.json", `{"server":{"addr":"second"}}`)
	loadConfig(t, &cfg, WithFile(first, second))
	if cfg.Server.Addr != "second" || cfg.Enabled || cfg.Count != 0 || cfg.Name != "" || cfg.Source != "default-source" {
		t.Fatalf("unexpected merged config: %+v", cfg)
	}
	// Dots separate nested fields; key case and string values are preserved.
	t.Setenv("COREX_NATIVE_server.addr", "env")
	loadConfig(t, &cfg, WithEnv("COREX_NATIVE_"), WithFile(first, second))
	if cfg.Server.Addr != "env" {
		t.Fatalf("environment did not override files: %+v", cfg)
	}
}

func TestLoadMatchesKratos(t *testing.T) {
	t.Setenv("COREX_NATIVE_ADDR", "from-env")
	t.Setenv("COREX_NATIVE_SERVER__ADDR", "not-a-nested-field")
	t.Setenv("COREX_NATIVE_EMPTY", "")
	path := writeConfig(t, "config.yaml", `server: {addr: "${ADDR:localhost}"}
name: "${MISSING:fallback}"
source: "${MISSING}"
extra: "${EMPTY:fallback}"
`)
	var got, want testConfig
	loadConfig(t, &got, WithFile(path), WithEnv("COREX_NATIVE_"))
	c := kratosconfig.New(kratosconfig.WithSource(file.NewSource(path), env.NewSource("COREX_NATIVE_")))
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	if err := c.Scan(&want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) || got.Server.Addr != "from-env" || got.Name != "fallback" || got.Source != "" {
		t.Fatalf("wrapper differs from Kratos: got %+v, want %+v", got, want)
	}
}

func TestLoadNativeTypeBehavior(t *testing.T) {
	t.Run("env values stay strings", func(t *testing.T) {
		t.Setenv("COREX_TYPES_count", "10")
		var cfg testConfig
		if closer, err := Load(&cfg, WithEnv("COREX_TYPES_")); err == nil {
			_ = closer.Close()
			t.Fatal("expected native JSON string-to-int decoding error")
		} else if closer != nil {
			t.Fatal("failed load returned a closer")
		}
	})
	t.Run("Go duration strings are not converted", func(t *testing.T) {
		var cfg struct {
			Timeout time.Duration `json:"timeout"`
		}
		path := writeConfig(t, "config.yaml", "timeout: 5s\n")
		if closer, err := Load(&cfg, WithFile(path)); err == nil {
			_ = closer.Close()
			t.Fatal("expected native duration decoding error")
		} else if closer != nil {
			t.Fatal("failed load returned a closer")
		}
	})
}

func TestLoadProtobuf(t *testing.T) {
	cfg := &descriptorpb.FileOptions{GoPackage: proto.String("default"), JavaMultipleFiles: proto.Bool(true)}
	path := writeConfig(t, "proto.yaml", "go_package: ${PACKAGE}\njava_multiple_files: false\noptimize_for: CODE_SIZE\n")
	t.Setenv("COREX_PROTO_PACKAGE", "from-env")
	loadConfig(t, cfg, WithFile(path), WithEnv("COREX_PROTO_"))
	if cfg.GetGoPackage() != "from-env" || cfg.GetJavaMultipleFiles() || cfg.GetOptimizeFor() != descriptorpb.FileOptions_CODE_SIZE {
		t.Fatalf("unexpected protobuf config: %v", cfg)
	}
}

func TestLoadProtobufDefaults(t *testing.T) {
	// FileDescriptorSet has nested protobuf messages; this also verifies defaults
	// survive a native protobuf Scan rather than relying on struct merging.
	cfg := &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{Name: proto.String("default.proto")}}}
	loadConfig(t, cfg)
	if cfg.GetFile()[0].GetName() != "default.proto" {
		t.Fatalf("protobuf defaults lost: %v", cfg)
	}
}

func TestLoadProtobufEnumDefaults(t *testing.T) {
	fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:    proto.String("config_enum_test.proto"),
		Package: proto.String("configtest"),
		Syntax:  proto.String("proto3"),
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("Mode"),
			Value: []*descriptorpb.EnumValueDescriptorProto{
				{Name: proto.String("MODE_UNSPECIFIED"), Number: proto.Int32(0)},
				{Name: proto.String("MODE_DEFAULT"), Number: proto.Int32(1)},
				{Name: proto.String("MODE_OTHER"), Number: proto.Int32(2)},
				{Name: proto.String("MODE_NEGATIVE"), Number: proto.Int32(-1)},
			},
		}},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Config"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("mode"),
				Number:   proto.Int32(1),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
				TypeName: proto.String(".configtest.Mode"),
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	desc := fd.Messages().Get(0)
	field := desc.Fields().ByName("mode")
	for _, tt := range []struct {
		name string
		yaml string
		want protoreflect.EnumNumber
	}{
		{name: "omitted preserves default", yaml: "{}\n", want: 1},
		{name: "numeric zero overrides default", yaml: "mode: 0\n", want: 0},
		{name: "named zero overrides default", yaml: "mode: MODE_UNSPECIFIED\n", want: 0},
		{name: "positive value overrides default", yaml: "mode: MODE_OTHER\n", want: 2},
		{name: "numeric negative overrides default", yaml: "mode: -1\n", want: -1},
		{name: "named negative overrides default", yaml: "mode: MODE_NEGATIVE\n", want: -1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := dynamicpb.NewMessage(desc)
			cfg.Set(field, protoreflect.ValueOfEnum(1))
			path := writeConfig(t, "enum.yaml", tt.yaml)
			loadConfig(t, cfg, WithFile(path))
			if got := cfg.Get(field).Enum(); got != tt.want {
				t.Fatalf("enum after Load = %d, want %d (default was 1)", got, tt.want)
			}
		})
	}
}

func TestLoadErrors(t *testing.T) {
	var cfg testConfig
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	if closer, err := Load(&cfg, WithFile(missing)); err == nil {
		_ = closer.Close()
		t.Fatal("expected file read error")
	} else if closer != nil {
		t.Fatal("failed load returned a closer")
	}
	path := writeConfig(t, "bad.yaml", "server: [\n")
	if closer, err := Load(&cfg, WithFile(path)); err == nil {
		_ = closer.Close()
		t.Fatal("expected YAML parse error")
	} else if closer != nil {
		t.Fatal("failed load returned a closer")
	}
}

func TestLoadObserver(t *testing.T) {
	path := writeConfig(t, "config.yaml", "name: initial\n")
	var cfg testConfig
	changes := make(chan string, 16)
	loadConfig(t, &cfg, WithFile(path), WithObserver("name", func(key string, value Value) {
		text, err := value.String()
		if err == nil && key == "name" {
			changes <- text
		}
	}))
	if cfg.Name != "initial" {
		t.Fatalf("initial config was not scanned: %+v", cfg)
	}
	select {
	case <-changes:
		t.Fatal("observer called for initial value")
	default:
	}
	if err := os.WriteFile(path, []byte("name: updated\n"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case text := <-changes:
		if text != "updated" {
			t.Fatalf("unexpected observer value: %q", text)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("observer did not receive file change")
	}
	if cfg.Name != "initial" {
		t.Fatal("observer unexpectedly changed destination")
	}
}

func TestLoadObserverUnknownKey(t *testing.T) {
	var cfg testConfig
	closer, err := Load(&cfg, WithObserver("missing", func(string, Value) {}))
	if err == nil {
		_ = closer.Close()
		t.Fatal("expected unknown observer key error")
	}
	if closer != nil {
		t.Fatal("failed observer registration returned a closer")
	}
}
