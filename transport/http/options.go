package http

import (
	"time"

	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/nexuer/corex/middleware"
	"github.com/nexuer/log"
)

type options struct {
	logger        log.Printer
	accept        []string
	maxBodyBytes  int64
	filters       []FilterFunc
	serverOptions []kratoshttp.ServerOption
}

// ServerOption configures an HTTP server.
type ServerOption func(*options)

// Logger sets the logger used by the default HTTP handlers.
// A nil logger uses the default logger.
func Logger(logger log.Printer) ServerOption {
	return func(o *options) { o.logger = logger }
}

// Address sets the listen address. If omitted, the framework default is used.
func Address(addr string) ServerOption {
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, kratoshttp.Address(addr))
	}
}

// Timeout sets the request timeout. Zero disables the framework timeout.
// If omitted, the framework default is used.
func Timeout(timeout time.Duration) ServerOption {
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, kratoshttp.Timeout(timeout))
	}
}

// MaxBodyBytes limits the request body size. Non-positive values disable the limit.
func MaxBodyBytes(size int64) ServerOption {
	return func(o *options) { o.maxBodyBytes = size }
}

// Accept sets the supported response media types, in preference order.
// An empty list uses JSON without content negotiation.
// Invalid media types are returned as errors by NewServer.
func Accept(types ...string) ServerOption {
	types = append([]string(nil), types...)
	return func(o *options) { o.accept = types }
}

// Filter appends HTTP filters in execution order.
func Filter(filters ...FilterFunc) ServerOption {
	filters = append([]FilterFunc(nil), filters...)
	return func(o *options) {
		o.filters = append(o.filters, filters...)
	}
}

// Middleware appends service middleware in execution order.
func Middleware(m ...middleware.Middleware) ServerOption {
	m = append([]middleware.Middleware(nil), m...)
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, kratoshttp.Middleware(m...))
	}
}
