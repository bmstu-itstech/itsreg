package apiv3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
)

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request, botID string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Commands.CreateRun.Handle(r.Context(), command.CreateRunRequest{
		ActorID: uid,
		BotID:   botID,
	})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, port.ErrActiveRunAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := CreateRunResponse{RunID: res.RunID}
	w.Header().Set("Content-Location", fmt.Sprintf("%s/runs/%s", s.prefix, res.RunID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, body)
}

func (s *Server) GetRuns(w http.ResponseWriter, r *http.Request, params GetRunsParams) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetRuns.Handle(r.Context(), query.GetRunsRequest{
		ActorID: uid,
		Status:  (*string)(params.Status),
		BotID:   params.BotID,
	})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := runsToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) GetBotRuns(w http.ResponseWriter, r *http.Request, id string, params GetBotRunsParams) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetBotRuns.Handle(r.Context(), query.GetBotRunsRequest{
		ActorID: uid,
		BotID:   id,
		Status:  (*string)(params.Status),
	})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := runsToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetRun.Handle(r.Context(), query.GetRunRequest{
		ActorID: uid,
		RunID:   id,
	})
	if errors.Is(err, port.ErrRunNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := ownedRunToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) StopRun(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	_, err := s.app.Commands.StopRun.Handle(r.Context(), command.StopRunRequest{
		ActorID: uid,
		RunID:   id,
	})
	if errors.Is(err, port.ErrRunNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, bots.ErrIllegalStateTransition) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Location", fmt.Sprintf("%s/runs/%s", s.prefix, id))
	w.WriteHeader(http.StatusAccepted)
}
