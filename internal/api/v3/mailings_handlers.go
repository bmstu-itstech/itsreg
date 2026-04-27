package apiv3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app/command"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/app/query"
	"github.com/bmstu-itstech/itsreg/internal/domain/shared"
	"github.com/go-chi/render"
)

func (s *Server) CreateMailing(w http.ResponseWriter, r *http.Request) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	var req CreateMailingRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := s.app.Commands.CreateMailing.Handle(r.Context(), command.CreateMailingRequest{
		ActorID:    uid,
		BotID:      req.BotID,
		Name:       req.Name,
		EntryKey:   req.EntryKey,
		Recipients: req.Recipients,
	})
	var vErr shared.ValidationError
	if errors.As(err, &vErr) {
		renderValidationError(w, r, vErr)
		return
	}
	if errors.Is(err, port.ErrMailingAlreadyExists) {
		renderPlainError(w, r, err, http.StatusConflict)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := CreateMailingResponse{MailingID: res.MailingID}
	w.Header().Set("Content-Location", fmt.Sprintf("%s/mailings/%s", s.prefix, res.MailingID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, body)
}

func (s *Server) GetMailing(w http.ResponseWriter, r *http.Request, id string) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetMailing.Handle(r.Context(), query.GetMailingRequest{
		ActorID:   uid,
		MailingID: id,
	})
	if errors.Is(err, port.ErrMailingNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := ownedMailingToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) GetMailings(w http.ResponseWriter, r *http.Request, params GetMailingsParams) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetMailings.Handle(r.Context(), query.GetMailingsRequest{
		ActorID: uid,
		BotID:   params.BotID,
		Status:  (*string)(params.Status),
	})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := mailingsToAPI(res)
	render.JSON(w, r, body)
}

func (s *Server) GetBotMailings(w http.ResponseWriter, r *http.Request, botID string, params GetBotMailingsParams) {
	uid, ok := jwtauth.FromContext(r.Context())
	if !ok {
		renderPlainError(w, r, ErrAuthorizationRequired, http.StatusUnauthorized)
		return
	}

	res, err := s.app.Queries.GetBotMailings.Handle(r.Context(), query.GetBotMailingsRequest{
		ActorID: uid,
		BotID:   botID,
		Status:  (*string)(params.Status),
	})
	if errors.Is(err, port.ErrMailingNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderInternalServerError(w, r)
		return
	}

	body := mailingsToAPI(res)
	render.JSON(w, r, body)
}
