package apiv3

import (
	"net/http"

	"github.com/bmstu-itstech/itsreg/internal/app"
	"github.com/go-chi/render"
)

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

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"status": "ok"})
}
