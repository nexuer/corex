package grpc

import "github.com/go-kratos/kratos/v2/transport/grpc"

type Server = grpc.Server
type ServerOption = grpc.ServerOption

var (
	NewServer = grpc.NewServer
)
