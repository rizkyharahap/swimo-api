package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"haphap/swimo-api/database"
	"haphap/swimo-api/internal/config"
	"haphap/swimo-api/internal/delivery/http/handler"
	"haphap/swimo-api/internal/delivery/http/router"
	"haphap/swimo-api/internal/domain/repository"
	"haphap/swimo-api/internal/domain/usecase"
)

func main() {
	// init logger
	_, cleanup, _ := config.InitLogging("swimo", getenv("LOG_LEVEL", "info"), getenv("LOG_FORMAT", "json"), getenv("LOG_FILE", ""), true)
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
	srv := config.NewServer(cfg)

	// repositories
	accountRepo := repository.NewAccountRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(db.Pool)

	// usecases
	signinUC := usecase.NewSignInUseCase(cfg, accountRepo, userRepo, sessionRepo)
	signinGuest := usecase.NewSignInGuestUseCase(cfg, sessionRepo)
	signupUC := usecase.NewSignUpUsecase(db.Pool, accountRepo, userRepo)

	// handlers
	authHandler := handler.NewAuthHandler(
		signinUC,
		signinGuest,
		signupUC,
	)

	// routes
	router.Register(srv.App, authHandler)

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
