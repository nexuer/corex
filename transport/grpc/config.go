package grpc

import (
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/nexuer/corex/log"
	"github.com/nexuer/corex/middleware"
)

type Config struct {
	Logger  *log.Logger
	Address string
	Timeout time.Duration
}

func (c *Config) Build(m ...middleware.Middleware) []ServerOption {
	return []ServerOption{
		grpc.Address(c.Address),
		grpc.Timeout(c.Timeout),
	}
}
