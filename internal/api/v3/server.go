package apiv3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bmstu-itstech/itsreg/internal/app/command"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
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

func (s *Server) DeleteScript(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	_, err := s.app.Commands.DeleteScript.Handle(r.Context(), command.DeleteScriptRequest{
		ActorID:  uid,
		ScriptID: id,
	})
	if errors.Is(err, port.ErrScriptNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	render.NoContent(w, r)
}

func (s *Server) GetScript(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetScript.Handle(r.Context(), query.GetScriptRequest{
		ActorID:  uid,
		ScriptID: id,
	})
	if errors.Is(err, port.ErrScriptNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := scriptToAPI(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, body)
}

func (s *Server) GetScripts(w http.ResponseWriter, r *http.Request) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetScripts.Handle(r.Context(), query.GetScriptsRequest{
		ActorID: uid,
	})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := scriptsToAPI(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, body)
}

func (s *Server) CreateBot(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetBot(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) CreateScript(w http.ResponseWriter, r *http.Request) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	var req CreateScriptRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := s.app.Commands.CreateScript.Handle(r.Context(), command.CreateScriptRequest{
		ActorID: uid,
		Desc:    req.Desc,
		Nodes:   nodesFromAPI(req.Nodes),
		Entries: entriesFromAPI(req.Entries),
	})
	if errors.Is(err, port.ErrScriptAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := CreateScriptResponse{ScriptID: res.ScriptID}
	w.Header().Set("Content-Location", fmt.Sprintf("%s/scripts/%s", s.prefix, res.ScriptID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, body)
}
