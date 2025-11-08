package sl

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
)

type watermillLoggerAdapter struct {
	log *slog.Logger
}

func NewWatermillLoggerAdapter(log *slog.Logger) watermill.LoggerAdapter {
	return &watermillLoggerAdapter{
		log: log,
	}
}

func (s watermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	fields = fields.Add(map[string]any{"error": err})
	s.log.Error(msg, s.attrs(fields)...)
}

func (s watermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	s.log.Info(msg, s.attrs(fields)...)
}

func (s watermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	s.log.Debug(msg, s.attrs(fields)...)
}

func (s watermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	s.log.Info(msg, s.attrs(fields)...)
}

func (s watermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return watermillLoggerAdapter{
		s.log.With(s.attrs(fields)...),
	}
}

func (s watermillLoggerAdapter) attrs(fields watermill.LogFields) []any {
	slogAttrs := make([]any, 0)
	for k, v := range fields {
		slogAttrs = append(slogAttrs, slog.Any(k, v))
	}
	return slogAttrs
}
