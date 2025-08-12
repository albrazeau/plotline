package api

import (
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *API) requestLogger(c *gin.Context) {
	start := time.Now()
	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	attrs := []slog.Attr{
		slog.String("method", c.Request.Method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Int64("latency_ms", latency.Milliseconds()),
		slog.String("ip", c.ClientIP()),
		slog.Int("size", c.Writer.Size()),
		slog.String("user_agent", c.Request.UserAgent()),
	}

	if reqID := c.GetHeader("X-Request-ID"); reqID != "" {
		attrs = append(attrs, slog.String("request_id", reqID))
	}
	if len(c.Errors) > 0 {
		attrs = append(attrs, slog.String("errors", c.Errors.String()))
	}

	switch {
	case status >= 500:
		a.logger.LogAttrs(c.Request.Context(), slog.LevelError, "http_request", attrs...)
	case status >= 400:
		a.logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "http_request", attrs...)
	default:
		a.logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "http_request", attrs...)
	}
}

func (a *API) recovery(c *gin.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			attrs := []slog.Attr{
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.String("ip", c.ClientIP()),
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			}
			if reqID := c.GetHeader("X-Request-ID"); reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			a.logger.LogAttrs(c.Request.Context(), slog.LevelError, "panic_recovered", attrs...)
			c.AbortWithStatus(500)
		}
	}()
	c.Next()
}
