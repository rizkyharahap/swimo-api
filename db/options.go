package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*pgxpool.Config)

func WithMaxConns(n int32) Option {
	return func(c *pgxpool.Config) { c.MaxConns = n }
}
func WithMinConns(n int32) Option {
	return func(c *pgxpool.Config) { c.MinConns = n }
}
func WithMaxConnLifetime(d time.Duration) Option {
	return func(c *pgxpool.Config) { c.MaxConnLifetime = d }
}
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(c *pgxpool.Config) { c.MaxConnIdleTime = d }
}
