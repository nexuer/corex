package sql

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/puddle/v2"
)

// No connections are precreated, so these tests do not require a database server.
const testSource = "postgres://test@127.0.0.1:1/test?sslmode=disable&pool_min_conns=0&pool_min_idle_conns=0&pool_max_conns=2"

func TestOpenClose(t *testing.T) {
	db, err := Open(Postgres, testSource)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if db.DB == nil || db.pool == nil {
		t.Fatal("missing SQL handle or PostgreSQL pool")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err == nil {
		t.Fatal("SQL handle still accepts work after Close")
	}
	conn, err := db.pool.Acquire(ctx)
	if conn != nil {
		conn.Release()
	}
	if !errors.Is(err, puddle.ErrClosedPool) {
		t.Fatalf("underlying pool was not closed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("repeated Close failed: %v", err)
	}
}

func TestCloseWithoutUnderlyingPool(t *testing.T) {
	handle, err := sql.Open("pgx", "postgres://test@127.0.0.1:1/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db := &DB{DB: handle}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.PingContext(context.Background()); err == nil {
		t.Fatal("SQL handle still accepts work after Close")
	}
}
