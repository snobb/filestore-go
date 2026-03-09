package pglib

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	PoolMinConns int32
	PoolMaxConns int32
	MaxConnIdle  time.Duration
}

func (c *Config) connString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		url.QueryEscape(c.Username),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database,
	)
}

var pgxpoolNewFunc = func(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, connString)
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if cfg.Host == "" {
		return nil, errors.New("database host cannot be blank")
	}

	if cfg.Port == 0 {
		cfg.Port = 5432
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.connString())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolCfg.MinConns = cfg.PoolMinConns
	poolCfg.MaxConns = cfg.PoolMaxConns
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdle

	pool, err := pgxpoolNewFunc(ctx, cfg.connString())
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func ParsePgUUID(pgUUID pgtype.UUID) (uuid.UUID, error) {
	if !pgUUID.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid uuid > missing")
	}

	var id uuid.UUID
	id, err := uuid.FromBytes(pgUUID.Bytes[:])
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unable to parse uuid %v > %w", pgUUID, err)
	}

	return id, nil
}
