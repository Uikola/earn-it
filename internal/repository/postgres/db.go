package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context, postgresConn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(postgresConn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	poolCfg.MinConns = 5
	poolCfg.MaxConns = 20
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute
	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
