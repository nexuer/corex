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
	MaxOpenConns int32
	// MaxIdleConns applies to standard SQL pools; it is not used by the PostgreSQL pool.
	MaxIdleConns    int32
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func pgxDB(dialect string, source string, c *Config) (*DB, error) {
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
	return &DB{DB: stdlib.OpenDBFromPool(pool), pool: pool}, nil
}

func parseConfig(dialect string, source string, c *Config) (*DB, error) {
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
		return &DB{DB: drv}, nil
	}
}

// DB owns a SQL handle and its underlying connection pool.
// Pass DB.DB to integrations that require *sql.DB, and close the owning DB
// to release all resources.
type DB struct {
	*sql.DB
	pool *pgxpool.Pool
}

// Close closes the SQL handle before closing its underlying pool.
func (db *DB) Close() error {
	err := db.DB.Close()
	if db.pool != nil {
		db.pool.Close()
	}
	return err
}

func Open(dialect string, source string, c ...*Config) (*DB, error) {
	if !validDriver(dialect) {
		return nil, fmt.Errorf("sql: unsupported driver %q", dialect)
	}
	var cfg *Config
	if len(c) > 0 {
		cfg = c[0]
	}

	drv, err := parseConfig(dialect, source, cfg)
	if err != nil {
		return nil, err
	}

	return drv, nil
}
