package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// App
	AppName string
	Env     string // dev|staging|prod

	// Logging
	LogLevel  string // debug|info|warn|error
	LogFormat string // json|text
	LogFile   string // path ke log file (kosong = stderr saja)
	LogAddSrc bool   // true untuk AddSource

	// Database
	DatabaseURL       string
	DBHost            string
	DBPort            int
	DBUser            string
	DBPass            string
	DBName            string
	DBSSLMode         string
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime time.Duration
	DBMaxConnIdleTime time.Duration
	DBHealthTimeout   time.Duration

	// HTTP / Fiber
	HTTPHost           string
	HTTPPort           int
	HTTPPrefork        bool
	HTTPReadTimeout    time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPIdleTimeout    time.Duration
	HTTPBodyLimitBytes int
	HTTPEnableETag     bool

	// CORS
	CORSAllowOrigins  string
	CORSAllowMethods  string
	CORSAllowHeaders  string
	CORSExposeHeaders string
	CORSCredentials   bool

	// Rate limit
	RateLimitEnabled   bool
	RateLimitMax       int
	RateLimitWindow    time.Duration
	RateLimitKeyHeader string

	JWTSecret     string        // minimal 32 chars
	JWTAccessTTL  time.Duration // ex: 15m
	JWTRefreshTTL time.Duration // ex: 720h (30d)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoiDef(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func Parse() *Config {
	cfg := &Config{
		// ===== App =====
		AppName: getEnv("APP_NAME", "swimo"),
		Env:     getEnv("APP_ENV", "dev"),

		// ===== Logging =====
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
		LogFile:   getEnv("LOG_FILE", ""),
		LogAddSrc: getEnv("LOG_ADD_SOURCE", "false") == "true",

		// ===== Database =====
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            atoiDef(getEnv("DB_PORT", "5432"), 5432),
		DBUser:            getEnv("DB_USER", "swimo_owner"),
		DBPass:            getEnv("DB_PASSWORD", "owner_pwd"),
		DBName:            getEnv("DB_NAME", "swimo"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxConns:        int32(atoiDef(getEnv("DB_MAX_CONNS", "15"), 15)),
		DBMinConns:        int32(atoiDef(getEnv("DB_MIN_CONNS", "2"), 2)),
		DBMaxConnLifetime: time.Duration(atoiDef(getEnv("DB_MAX_CONN_LIFETIME_SEC", "3600"), 3600)) * time.Second,
		DBMaxConnIdleTime: time.Duration(atoiDef(getEnv("DB_MAX_CONN_IDLE_SEC", "300"), 300)) * time.Second,
		DBHealthTimeout:   time.Duration(atoiDef(getEnv("DB_HEALTH_TIMEOUT_MS", "1500"), 1500)) * time.Millisecond,

		// ===== HTTP / Fiber =====
		HTTPHost:           getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:           atoiDef(getEnv("HTTP_PORT", "8080"), 8080),
		HTTPPrefork:        getEnv("HTTP_PREFORK", "false") == "true",
		HTTPReadTimeout:    time.Duration(atoiDef(getEnv("HTTP_READ_TIMEOUT_MS", "10000"), 10000)) * time.Millisecond,
		HTTPWriteTimeout:   time.Duration(atoiDef(getEnv("HTTP_WRITE_TIMEOUT_MS", "10000"), 10000)) * time.Millisecond,
		HTTPIdleTimeout:    time.Duration(atoiDef(getEnv("HTTP_IDLE_TIMEOUT_MS", "60000"), 60000)) * time.Millisecond,
		HTTPBodyLimitBytes: atoiDef(getEnv("HTTP_BODY_LIMIT_BYTES", "10485760"), 10<<20), // 10MB
		HTTPEnableETag:     getEnv("HTTP_ETAG", "true") == "true",

		// ===== CORS =====
		CORSAllowOrigins:  getEnv("CORS_ALLOW_ORIGINS", "*"),
		CORSAllowMethods:  getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		CORSAllowHeaders:  getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
		CORSExposeHeaders: getEnv("CORS_EXPOSE_HEADERS", ""),
		CORSCredentials:   getEnv("CORS_CREDENTIALS", "false") == "true",

		// ===== Rate limit =====
		RateLimitEnabled:   getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		RateLimitMax:       atoiDef(getEnv("RATE_LIMIT_MAX", "120"), 120),
		RateLimitWindow:    time.Duration(atoiDef(getEnv("RATE_LIMIT_WINDOW_SEC", "60"), 60)) * time.Second,
		RateLimitKeyHeader: getEnv("RATE_LIMIT_KEY_HEADER", ""),

		JWTSecret:     getEnv("JWT_SECRET", "dev-please-change-this-32chars-minimum"),
		JWTAccessTTL:  time.Duration(atoiDef(getEnv("JWT_ACCESS_TTL_MIN", "15"), 15)) * time.Minute,
		JWTRefreshTTL: time.Duration(atoiDef(getEnv("JWT_REFRESH_TTL_HOURS", "720"), 720)) * time.Hour,
	}

	// Build DATABASE_URL jika kosong
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
		)
	}

	// Log config (pakai slog)
	slog.Debug("config loaded",
		slog.String("env", cfg.Env),
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
		slog.String("log_file", cfg.LogFile),
		slog.String("db_host", cfg.DBHost),
		slog.Int("db_port", cfg.DBPort),
		slog.Int("http_port", cfg.HTTPPort),
		slog.Bool("http_prefork", cfg.HTTPPrefork),
		slog.String("cors_origins", cfg.CORSAllowOrigins),
		slog.Bool("ratelimit", cfg.RateLimitEnabled),
	)

	return cfg
}
