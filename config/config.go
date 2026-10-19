// Package config loads startup configuration from files and environment variables.
package config

import (
	"encoding/json"
	"errors"
	"io"

	kratosconfig "github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type options struct {
	files     []string
	prefixes  []string
	observers []observation
}

// Value provides access to a configuration value.
type Value = kratosconfig.Value

// Observer receives changes to a configuration value.
type Observer = kratosconfig.Observer

type observation struct {
	key      string
	observer Observer
}

// Option selects configuration sources.
type Option func(*options)

// WithFile adds file or directory sources in order.
func WithFile(paths ...string) Option {
	paths = append([]string(nil), paths...)
	return func(o *options) { o.files = append(o.files, paths...) }
}

// WithEnv adds an environment source using prefix and key semantics.
// Keys retain their case; dots separate nested fields, not double underscores.
// For example, APP_ADDR is available as ADDR in ${ADDR:default} with prefix APP_.
// Values remain strings; no custom type conversion is performed.
func WithEnv(prefix string) Option {
	return func(o *options) { o.prefixes = append(o.prefixes, prefix) }
}

// WithObserver registers a callback for changes to an existing configuration key.
// It does not invoke the callback for the initial value or update dst automatically.
// Callbacks run on source watcher goroutines and must synchronize shared state.
func WithObserver(key string, observer Observer) Option {
	return func(o *options) {
		o.observers = append(o.observers, observation{key: key, observer: observer})
	}
}

// Load loads dst's existing defaults, then files, then environment sources.
// Unknown fields are ignored; missing placeholders without defaults become empty
// strings. Failed decoding may modify dst. Protobuf defaults use proto field names, which
// should also be used in sources to avoid duplicate snake_case/camelCase keys.
// On success, the caller must close the returned handle to release watchers.
// On failure, resources are closed internally and the returned handle is nil.
func Load(dst any, opts ...Option) (_ io.Closer, err error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	var data []byte
	if message, ok := dst.(proto.Message); ok {
		data, err = (protojson.MarshalOptions{UseProtoNames: true}).Marshal(message)
	} else {
		data, err = json.Marshal(dst)
	}
	if err != nil {
		return nil, err
	}
	sources := []kratosconfig.Source{defaultsSource{data: data}}
	for _, path := range o.files {
		sources = append(sources, file.NewSource(path))
	}
	for _, prefix := range o.prefixes {
		sources = append(sources, env.NewSource(prefix))
	}
	c := kratosconfig.New(kratosconfig.WithSource(sources...))
	defer func() {
		if err != nil {
			err = errors.Join(err, c.Close())
		}
	}()
	if err = c.Load(); err != nil {
		return nil, err
	}
	for _, observation := range o.observers {
		if err = c.Watch(observation.key, observation.observer); err != nil {
			return nil, err
		}
	}
	if err = c.Scan(dst); err != nil {
		return nil, err
	}
	return c, nil
}

// defaultsSource exposes existing values as a JSON configuration source.
type defaultsSource struct {
	data []byte
}

func (s defaultsSource) Load() ([]*kratosconfig.KeyValue, error) {
	return []*kratosconfig.KeyValue{{Key: "defaults", Format: "json", Value: s.data}}, nil
}

func (defaultsSource) Watch() (kratosconfig.Watcher, error) {
	return env.NewWatcher()
}
