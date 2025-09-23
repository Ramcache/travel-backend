package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	DSN         string
	MaxConns    int32
	MinConns    int32
	ConnTimeout time.Duration
	IdleTimeout time.Duration
}

// NewPostgres создаёт и настраивает пул соединений
func NewPostgres(ctx context.Context, cfg PostgresConfig) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MaxConns > 0 {
		pgxCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		pgxCfg.MinConns = cfg.MinConns
	}
	if cfg.IdleTimeout > 0 {
		pgxCfg.MaxConnIdleTime = cfg.IdleTimeout
	}
	if cfg.ConnTimeout > 0 {
		pgxCfg.ConnConfig.ConnectTimeout = cfg.ConnTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("pgx connect: %w", err)
	}

	return pool, nil
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
