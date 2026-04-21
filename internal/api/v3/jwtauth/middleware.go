package jwtauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5/request"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (bots.UserID, error)
}

type Middleware struct {
	verifier  TokenVerifier
	extractor request.Extractor
}

func NewMiddleware(verifier TokenVerifier) *Middleware {
	return &Middleware{
		verifier: verifier,
		extractor: request.MultiExtractor{
			request.AuthorizationHeaderExtractor,
			request.ArgumentExtractor{"jwtToken"},
		},
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := m.extractor.ExtractToken(r)
		if errors.Is(err, request.ErrNoTokenInRequest) {
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]interface{}{"message": err.Error()})
			return
		}

		uid, err := m.verifier.VerifyToken(r.Context(), tokenString)
		if errors.Is(err, port.ErrTokenInvalid) {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]interface{}{"message": fmt.Sprintf("token is invalid: %s", err.Error())})
			return
		} else if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]interface{}{"message": "internal server error"})
		} else if uid == 0 {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]interface{}{"message": "token is invalid: uid can't be empty"})
			return
		}

		r = r.WithContext(toContext(r.Context(), uid.Int64()))

		next.ServeHTTP(w, r)
	})
}
