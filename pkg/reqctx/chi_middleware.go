package reqctx

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func ChiRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID := middleware.GetReqID(r.Context()); requestID != "" {
			r = r.WithContext(WithRequestID(r.Context(), requestID))
		}

		next.ServeHTTP(w, r)
	})
}
