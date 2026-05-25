package storage

import (
	"context"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg config.Config) (*pgxpool.Pool, error) {
	connStr := "user=" + cfg.DBUser + " password=" + cfg.DBPassword + " host=" + cfg.DBHost + " port=" + cfg.DBPort + " dbname=" + cfg.DBName
	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
