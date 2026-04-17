package apiv3

import (
	"errors"
	"fmt"
	"net/http"

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
