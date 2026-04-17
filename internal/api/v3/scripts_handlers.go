package apiv3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

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
	var vErr shared.ValidationError
	if errors.As(err, &vErr) {
		renderValidationError(w, r, vErr)
		return
	}
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
	render.JSON(w, r, body)
}

func (s *Server) UpdateScript(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	var req Script
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := s.app.Commands.UpdateScript.Handle(r.Context(), command.UpdateScriptRequest{
		ActorID:  uid,
		ScriptID: id,
		Desc:     req.Desc,
		Nodes:    nodesFromAPI(req.Nodes),
		Entries:  entriesFromAPI(req.Entries),
	})
	var vErr shared.ValidationError
	if errors.As(err, &vErr) {
		renderValidationError(w, r, vErr)
		return
	}
	if errors.Is(err, port.ErrScriptNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := scriptToAPI(res)
	render.JSON(w, r, body)
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
