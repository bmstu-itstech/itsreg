package jwtauth

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
)

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claims accessTokenClaims
		token, err := request.ParseFromRequest(
			r,
			request.MultiExtractor{
				request.AuthorizationHeaderExtractor,
				request.ArgumentExtractor{"jwtToken"},
			},
			func(_ *jwt.Token) (_ interface{}, _ error) {
				return []byte(secret), nil
			},
			request.WithClaims(&claims),
		)
		if err != nil {
			httpError(w, r, err, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			httpError(w, r, errors.New("invalid JWT token"), http.StatusBadRequest)
			return
		}

		r = r.WithContext(userUUIDToContext(r.Context(), claims.UserUUID))

		next.ServeHTTP(w, r)
	})
}

type authorizationError struct {
	Message string `json:"message"`
}

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	w.WriteHeader(code)
	render.JSON(w, r, authorizationError{Message: err.Error()})
}
