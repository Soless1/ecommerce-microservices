package app

import (
	"context"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newPostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConstructPostgresURL())
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL pool configuration: %w", err)
	}
	poolConfig.MaxConns = cfg.PG.MaxConns
	poolConfig.MinConns = cfg.PG.MinConns
	poolConfig.HealthCheckPeriod = cfg.PG.HealthCheckPeriod
	poolConfig.MaxConnLifetime = cfg.PG.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.PG.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return pool, nil
}
