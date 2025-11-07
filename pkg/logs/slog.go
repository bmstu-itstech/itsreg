package logs

import (
	"log/slog"
	"os"

	"github.com/bmstu-itstech/itsreg-bots/pkg/logs/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var defaultLogger *slog.Logger

func DefaultLogger() *slog.Logger {
	if defaultLogger == nil {
		env := os.Getenv("APP_ENV")
		defaultLogger = NewLogger(env)
	}
	return defaultLogger
}

func NewLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envProd:
		log = slog.New(
			slogpretty.PrettyHandlerOptions{
				SlogOpts: &slog.HandlerOptions{
					Level: slog.LevelInfo,
				},
			}.NewPrettyHandler(os.Stdout),
		)
	case envLocal, envDev:
		log = slog.New(
			slogpretty.PrettyHandlerOptions{
				SlogOpts: &slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			}.NewPrettyHandler(os.Stdout),
		)
	default:
		log = slog.New(
			slogpretty.PrettyHandlerOptions{
				SlogOpts: &slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			}.NewPrettyHandler(os.Stdout),
		)
	}

	return log
}
