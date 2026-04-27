package logs

import (
	"log/slog"
	"os"

	"github.com/bmstu-itstech/itsreg/pkg/logs/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func NewLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
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
