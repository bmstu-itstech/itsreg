package apiv3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
)

func (s *Server) CreateBot(w http.ResponseWriter, r *http.Request) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	var req CreateBotRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := s.app.Commands.CreateBot.Handle(r.Context(), command.CreateBotRequest{
		ActorID:  uid,
		ScriptID: req.ScriptID,
		Token:    req.Token,
		Desc:     req.Desc,
	})
	var vErr shared.ValidationError
	if errors.As(err, &vErr) {
		renderValidationError(w, r, vErr)
		return
	}
	if errors.Is(err, port.ErrBotAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if errors.Is(err, port.ErrTokenAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if errors.Is(err, port.ErrScriptNotFound) {
		renderPlainError(w, r, err, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := CreateBotResponse{BotID: res.BotID}
	w.Header().Set("Content-Location", fmt.Sprintf("%s/bots/%s", s.prefix, res.BotID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, body)
}

func (s *Server) GetBots(w http.ResponseWriter, r *http.Request) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetBots.Handle(r.Context(), query.GetBotsRequest{ActorID: uid})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := botsToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) GetBot(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetBot.Handle(r.Context(), query.GetBotRequest{
		ActorID: uid,
		BotID:   id,
	})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := botToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) UpdateBot(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	var req UpdateBotRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := s.app.Commands.UpdateBot.Handle(r.Context(), command.UpdateBotRequest{
		ActorID:  uid,
		BotID:    id,
		ScriptID: req.ScriptID,
		Token:    req.Token,
		Desc:     req.Desc,
	})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, port.ErrTokenAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if errors.Is(err, port.ErrScriptNotFound) {
		renderPlainError(w, r, err, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := botToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) DeleteBot(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	_, err := s.app.Commands.DeleteBot.Handle(r.Context(), command.DeleteBotRequest{
		ActorID: uid,
		BotID:   id,
	})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	render.NoContent(w, r)
}
