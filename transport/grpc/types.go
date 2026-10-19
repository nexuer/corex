package grpc

import "github.com/go-kratos/kratos/v3/transport/grpc"

type Server = grpc.Server
type ServerOption = grpc.ServerOption

var (
	Address           = grpc.Address
	Timeout           = grpc.Timeout
	Middleware        = grpc.Middleware
	StreamMiddleware  = grpc.StreamMiddleware
	Network           = grpc.Network
	Endpoint          = grpc.Endpoint
	TLSConfig         = grpc.TLSConfig
	Listener          = grpc.Listener
	UnaryInterceptor  = grpc.UnaryInterceptor
	StreamInterceptor = grpc.StreamInterceptor
	CustomHealth      = grpc.CustomHealth
	DisableReflection = grpc.DisableReflection
	Options           = grpc.Options
)

// NewServer creates a gRPC server with the same constructor signature as HTTP.
// Construction currently always succeeds and returns a nil error.
func NewServer(opts ...ServerOption) (*Server, error) {
	return grpc.NewServer(opts...), nil
}
