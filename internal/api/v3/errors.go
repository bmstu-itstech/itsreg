package apiv3

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

var ErrAuthorizationRequired = errors.New("authorization required")

func renderPlainError(w http.ResponseWriter, r *http.Request, inner error, code int) {
	e := PlainError{Message: inner.Error()}
	render.Status(r, code)
	render.JSON(w, r, e)
}

func renderInternalServerError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("internal server error"))
}
