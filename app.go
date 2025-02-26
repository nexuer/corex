package corex

import "github.com/go-kratos/kratos/v2"

type App = kratos.App

var (
	New     = kratos.New
	Server  = kratos.Server
	Context = kratos.Context
	Name    = kratos.Name
)

var Registrar = kratos.Registrar
