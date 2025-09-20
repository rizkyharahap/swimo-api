package server

import (
	"haphap/swimo-api/config"
	"haphap/swimo-api/internal/middleware"
	"haphap/swimo-api/pkg/response"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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

			return c.Status(code).JSON(response.Base{
				Message: httpStatusMessage(code),
			})
		},
	})

	// Middlewares
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(requestid.New())

	// IP-based rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 60 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(response.Base{
				Message: "Too many requests, please try again later.",
			})
		},
	}))

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

	// Cache with revalidation
	app.Use(cache.New(cache.Config{
		Expiration:   600 * time.Second, // Cache TTL set to 600 seconds (10 minutes)
		CacheControl: true,              // Automatically sets Cache-Control header
	}))

	// Request logger (custom, using slog)
	app.Use(middleware.LoggingMiddleware)

	// Health & readiness
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(response.Base{Message: "ok"})
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.JSON(response.Base{Message: "ready"})
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
