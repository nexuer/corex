package http

import "github.com/go-kratos/kratos/v3/transport/http"

const (
	SupportPackageIsVersion1 = true
)

type Server = http.Server
type Context = http.Context
type FilterFunc = http.FilterFunc
type ServerStream = http.ServerStream
type ServerStreamOption = http.ServerStreamOption

var (
	SetOperation                   = http.SetOperation
	BodyContentType                = http.BodyContentType
	WithStreamBodyField            = http.WithStreamBodyField
	NewServerSentEventServerStream = http.NewServerSentEventServerStream
	NewWebSocketServerStream       = http.NewWebSocketServerStream
)
