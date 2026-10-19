package corex

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v3"
	internalruntime "github.com/nexuer/corex/internal/runtime"
	"github.com/nexuer/corex/transport"
	"github.com/nexuer/log"
	"github.com/zdz1715/appversion"
)

type options struct {
	ctx         context.Context
	id          string
	name        string
	version     string
	logger      *log.Logger
	stopTimeout time.Duration
	servers     []transport.Server
}

// Option configures an App.
type Option func(*options)

// Context sets the base application context.
func Context(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}
func ID(id string) Option {
	return func(o *options) { o.id = id }
}
func Name(name string) Option {
	return func(o *options) { o.name = name }
}
func Version(version string) Option {
	return func(o *options) { o.version = version }
}
func StopTimeout(timeout time.Duration) Option {
	return func(o *options) { o.stopTimeout = timeout }
}

// Logger sets the Nex logger used by the underlying runtime.
func Logger(logger *log.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

// Server registers transport servers with the application.
func Server(servers ...transport.Server) Option {
	return func(o *options) { o.servers = append(o.servers, servers...) }
}

// App manages the lifecycle of a application.
//
// Kratos is the current runtime implementation. Keeping it behind App leaves
// the entrypoint stable without duplicating its lifecycle implementation.
type App struct {
	app *kratos.App
}

// New creates an application lifecycle manager.
func New(opts ...Option) *App {
	o := options{
		ctx:     context.Background(),
		logger:  log.Default(),
		version: appversion.Get().Version,
	}
	for _, opt := range opts {
		opt(&o)
	}

	kratosOpts := []kratos.Option{
		kratos.Context(o.ctx),
		kratos.Logger(internalruntime.Logger(o.logger)),
		kratos.Server(o.servers...),
	}
	if o.id != "" {
		kratosOpts = append(kratosOpts, kratos.ID(o.id))
	}
	if o.name != "" {
		kratosOpts = append(kratosOpts, kratos.Name(o.name))
	}
	if o.version != "" {
		kratosOpts = append(kratosOpts, kratos.Version(o.version))
	}
	if o.stopTimeout > 0 {
		kratosOpts = append(kratosOpts, kratos.StopTimeout(o.stopTimeout))
	}
	app := kratos.New(kratosOpts...)
	return &App{app: app}
}

// Run starts the application and blocks until it stops.
func (a *App) Run() error {
	return a.app.Run()
}

// Stop gracefully stops the application.
func (a *App) Stop() error {
	return a.app.Stop()
}

func (a *App) ID() string      { return a.app.ID() }
func (a *App) Name() string    { return a.app.Name() }
func (a *App) Version() string { return a.app.Version() }
