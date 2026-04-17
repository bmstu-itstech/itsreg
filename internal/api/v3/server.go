package apiv3

import "github.com/bmstu-itstech/itsreg/internal/app"

type Server struct {
	app    *app.Application
	prefix string
}

func NewServer(a *app.Application, prefix string) *Server {
	return &Server{
		app:    a,
		prefix: prefix,
	}
}
