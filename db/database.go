package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string, opts ...Option) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("db parse config failed", slog.String("err", err.Error()))
		return nil, err
	}

	// defaults
	cfg.MaxConns = 15
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 5 * time.Minute

	for _, opt := range opts {
		opt(cfg)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("db pool create failed", slog.String("err", err.Error()))
		return nil, err
	}

	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err != nil {
		slog.Error("db ping failed", slog.String("err", err.Error()))
		pool.Close()
		return nil, err
	}

	slog.Info("db connected")
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		slog.Info("db closing")
		db.Pool.Close()
	}
}
