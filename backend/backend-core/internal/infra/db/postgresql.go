package db

import (
	"context"
	"fmt"
	"time"

	"pinterest-core/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*PostgreSQL, error) {
	tctx, cancel := context.WithTimeout(ctx, cfg.PostgresTimeout())
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.MaxConns = cfg.PostgresMaxConns
	poolCfg.MinConns = cfg.PostgresMinConns
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(tctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(tctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgreSQL{Pool: pool}, nil
}

func (p *PostgreSQL) Close(context.Context) error {
	if p == nil || p.Pool == nil {
		return nil
	}

	p.Pool.Close()
	return nil
}
