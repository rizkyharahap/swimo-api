package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"haphap/swimo-api/config"
	"haphap/swimo-api/database"
	"haphap/swimo-api/internal/app/auth"
	"haphap/swimo-api/internal/app/auth/delivery/http"
	"haphap/swimo-api/internal/server"
	"haphap/swimo-api/pkg/logging"
)

func main() {
	// init logger
	_, cleanup, _ := logging.Init("swimo", getenv("LOG_LEVEL", "info"), getenv("LOG_FORMAT", "json"), getenv("LOG_FILE", ""), true)
	defer cleanup()

	// config
	cfg := config.Parse()

	// database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		slog.Error("database connect failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// server
	srv := server.NewServer(cfg)

	// repositories
	authRepo := auth.NewAuthRepository(db.Pool)

	// usecases
	authUsecase := auth.NewAuthUseCase(cfg, db.Pool, authRepo)

	// handlers
	authHandler := http.NewAuthHandler(authUsecase)

	// routes
	http.Register(srv.App, authHandler)

	// run + graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Listen(cfg)
	}()

	// wait for signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		slog.Info("signal received", slog.String("sig", sig.String()))
	case err := <-errCh:
		if err != nil {
			slog.Error("fiber listen error", slog.String("err", err.Error()))
		}
	}

	_ = srv.Shutdown()
	time.Sleep(150 * time.Millisecond)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
