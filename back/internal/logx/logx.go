package logx

import (
	"log/slog"
	"os"
	"time"
)

func InitLogger(isProd bool) *slog.Logger {

	level := slog.LevelDebug
	if isProd {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339Nano))
			}
			return a
		},
	}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if isProd {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
