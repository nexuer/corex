package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	Postgres = "postgres"
)

var drivers = map[string]string{
	Postgres: "pgx",
}

func validDriver(driver string) bool {
	_, ok := drivers[driver]
	return ok
}

type Config struct {
	MaxOpenConns    int32
	MaxIdleConns    int32
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func pgxDB(dialect string, source string, c *Config) (*sql.DB, error) {
	poolCfg, err := pgxpool.ParseConfig(source)
	if err != nil {
		return nil, err
	}

	if c != nil {
		poolCfg.MaxConns = c.MaxOpenConns
		poolCfg.MaxConnLifetime = c.ConnMaxLifetime
		poolCfg.MaxConnIdleTime = c.ConnMaxIdleTime
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDBFromPool(pool), nil
}

func parseConfig(dialect string, source string, c *Config) (*sql.DB, error) {
	switch drivers[dialect] {
	case "pgx":
		return pgxDB(dialect, source, c)
	default:
		drv, err := sql.Open(dialect, source)
		if err != nil {
			return nil, err
		}
		if c != nil {
			drv.SetMaxOpenConns(int(c.MaxOpenConns))
			drv.SetMaxIdleConns(int(c.MaxIdleConns))
			drv.SetConnMaxLifetime(c.ConnMaxLifetime)
			drv.SetConnMaxIdleTime(c.ConnMaxIdleTime)
		}
		return drv, nil
	}
}

func Open(dialect string, source string, c *Config) (*sql.DB, error) {
	if !validDriver(dialect) {
		return nil, fmt.Errorf("sql: unsupported driver %q", dialect)
	}

	drv, err := parseConfig(dialect, source, c)
	if err != nil {
		return nil, err
	}

	return drv, nil
}
