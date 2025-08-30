package logx

import (
	"log/slog"
	"main/config"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// InitLogger creates a slog.Logger with the configured level and format,
// applies consistent time formatting, and sets it as the global default logger.
func InitLogger(logCfg config.LogConfig) *slog.Logger {
	level := parseLogLevel(logCfg.Level)

	var handler slog.Handler
	if logCfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
					a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339Nano))
				}
				return a
			},
			AddSource: true,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: time.RFC3339Nano,
			AddSource:  true,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
