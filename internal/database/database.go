package database

import (
	"context"
	"fmt"
	"log"

	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("connect error: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	log.Println("✓ Database connected")
	return pool, nil
}
