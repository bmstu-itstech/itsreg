package sl

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const elapsedPrecision = time.Millisecond / 100

type chiLoggerMiddleware struct {
	log *slog.Logger
}

func NewLoggerMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&chiLoggerMiddleware{log})
}

func (m *chiLoggerMiddleware) NewLogEntry(r *http.Request) middleware.LogEntry {
	log := m.log.With(
		slog.String("req_id", middleware.GetReqID(r.Context())),
		slog.String("http_method", r.Method),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("uri", r.RequestURI),
	)
	log.InfoContext(r.Context(), "request started")

	return &chiLoggerEntry{log: log}
}

func (m *chiLoggerMiddleware) GetLogEntry(r *http.Request) middleware.LogEntry {
	entry, _ := middleware.GetLogEntry(r).(*chiLoggerEntry)
	return entry
}

type chiLoggerEntry struct {
	log *slog.Logger
}

func (l *chiLoggerEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	l.log = l.log.With(
		slog.Int("resp_status", status),
		slog.Int("resp_bytes_length", bytes),
		slog.String("resp_elapsed", elapsed.Round(elapsedPrecision).String()),
	)

	l.log.Info("request completed")
}

func (l *chiLoggerEntry) Panic(v interface{}, stack []byte) {
	l.log = l.log.With(
		slog.String("stack", string(stack)),
		slog.String("panic", fmt.Sprintf("%+v", v)),
	)
}
