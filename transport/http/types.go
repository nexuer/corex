package http

import "github.com/go-kratos/kratos/v2/transport/http"

const (
	SupportPackageIsVersion1 = true
)

type Server = http.Server
type ServerOption = http.ServerOption
type Context = http.Context
type CallOption = http.CallOption

var (
	SetOperation = http.SetOperation
	NewServer    = http.NewServer
)
