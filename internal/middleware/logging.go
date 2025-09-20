package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func LoggingMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	lat := time.Since(start)

	attrs := []any{
		slog.String("rid", string(c.Response().Header.Peek("X-Request-ID"))),
		slog.String("method", c.Method()),
		slog.String("path", c.OriginalURL()),
		slog.Int("status", c.Response().StatusCode()),
		slog.String("ip", c.IP()),
		slog.Duration("latency", lat),
		slog.Int("bytes_in", len(c.Request().Body())),
		slog.Int("bytes_out", len(c.Response().Body())),
	}

	// Log level by status
	status := c.Response().StatusCode()
	switch {
	case status >= 500:
		slog.Error("http", attrs...)
	case status >= 400:
		slog.Warn("http", attrs...)
	default:
		slog.Info("http", attrs...)
	}
	return err
}
