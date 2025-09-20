package server

import (
	"haphap/swimo-api/config"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Server struct {
	App *fiber.App
}

func NewServer(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		Prefork:       cfg.HTTP.Prefork,
		ReadTimeout:   cfg.HTTP.ReadTimeout,
		WriteTimeout:  cfg.HTTP.WriteTimeout,
		IdleTimeout:   cfg.HTTP.IdleTimeout,
		BodyLimit:     cfg.HTTP.BodyLimitBytes,
		CaseSensitive: true,
		StrictRouting: false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			slog.Error("http error",
				slog.Int("status", code),
				slog.String("method", c.Method()),
				slog.String("path", c.OriginalURL()),
				slog.String("err", err.Error()),
			)

			return c.Status(code).JSON(fiber.Map{
				"message": httpStatusMessage(code),
			})
		},
	})

	// Middlewares
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(requestid.New())

	// CORS
	app.Use(cors.New(cors.Config{

		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.Credentials,
		MaxAge:           600, // seconds
	}))

	// ETag
	if cfg.HTTP.EnableETag {
		app.Use(etag.New())
	}

	// Compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))

	// Request logger (custom, using slog)
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		lat := time.Since(start)

		attrs := []any{
			slog.String("rid", string(c.Response().Header.Peek("X-Request-ID"))),
			slog.String("method", c.Method()),
			slog.String("path", c.OriginalURL()),
			slog.Int("status", c.Response().StatusCode()),
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
	})

	// Health & readiness
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ready"})
	})

	return &Server{App: app}
}

func (s *Server) Listen(cfg *config.Config) error {
	addr := net.JoinHostPort(cfg.HTTP.Host, strconv.Itoa(cfg.HTTP.Port))
	slog.Info("fiber listening", slog.String("addr", addr))
	return s.App.Listen(addr)
}

func (s *Server) Shutdown() error {
	slog.Info("fiber shutting down")
	return s.App.Shutdown()
}

func httpStatusMessage(code int) string {
	switch code {
	case fiber.StatusUnauthorized:
		return "Unauthorized"
	case fiber.StatusForbidden:
		return "Forbidden"
	case fiber.StatusNotFound:
		return "Not Found"
	case fiber.StatusTooManyRequests:
		return "Too Many Requests"
	case fiber.StatusBadRequest:
		return "Bad Request"
	case fiber.StatusUnprocessableEntity:
		return "Unprocessable Entity"
	default:
		if code >= 500 {
			return "Internal Server Error"
		}
		return "Error"
	}
}
