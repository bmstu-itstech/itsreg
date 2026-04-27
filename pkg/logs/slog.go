package logs

import (
	"log/slog"
	"os"

	"github.com/bmstu-itstech/itsreg/pkg/logs/handlers/slogcontext"
	"github.com/bmstu-itstech/itsreg/pkg/logs/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func NewLogger(env string) *slog.Logger {
	var log *slog.Logger

	var handler slog.Handler
	switch env {
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	case envLocal, envDev:
		handler = slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}.NewPrettyHandler(os.Stdout)
	default:
		handler = slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}.NewPrettyHandler(os.Stdout)
	}

	handler = slogcontext.NewContextHandler(handler)
	log = slog.New(handler)

	return log
}
