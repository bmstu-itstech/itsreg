package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func renderInvalidInputError(w http.ResponseWriter, r *http.Request, iiErr bots.InvalidInputError, code int) {
	var e Error
	err := e.FromInvalidInputError(InvalidInputError{
		Code:    iiErr.Code,
		Details: nilOnEmptyMap(iiErr.Details),
		Message: iiErr.Message,
	})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}
	render.Status(r, code)
	render.JSON(w, r, e)
}

func renderMultiError(w http.ResponseWriter, r *http.Request, mErr *bots.MultiError, code int) {
	var err2items Error2
	for _, err := range mErr.Errors {
		var iiErr bots.InvalidInputError
		var item Error_2_Item
		if errors.As(err, &iiErr) {
			mapErr := item.FromInvalidInputError(InvalidInputError{
				Code:    iiErr.Code,
				Details: nilOnEmptyMap(iiErr.Details),
				Message: iiErr.Message,
			})
			if mapErr != nil {
				renderInternalServerError(w, r)
				return
			}
		} else {
			mapErr := item.FromPlainError(PlainError{
				Message: err.Error(),
			})
			if mapErr != nil {
				renderInternalServerError(w, r)
				return
			}
		}
		err2items = append(err2items, item)
	}
	var e Error
	mapErr := e.FromError2(err2items)
	if mapErr != nil {
		renderInternalServerError(w, r)
		return
	}
	render.Status(r, code)
	render.JSON(w, r, e)
}

func renderPlainError(w http.ResponseWriter, r *http.Request, inner error, code int) {
	var e Error
	err := e.FromPlainError(PlainError{
		Message: inner.Error(),
	})
	if err != nil {
		renderInternalServerError(w, r)
		return
	}
	render.Status(r, code)
	render.JSON(w, r, e)
}

func renderInternalServerError(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("internal server error"))
}
