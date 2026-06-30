// Package store is byakugan's door to the corpus warehouse: Postgres + pgvector.
// It hides SQL behind typed methods so the rest of the brain never hand-writes a
// query. Phase B slice 1: connect + migrate. Upsert/Search land in later slices.
package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// initSchema is the migration SQL compiled straight into the binary. //go:embed
// reads the file at BUILD time and assigns its contents to the string below — so
// the shipped binary needs no loose .sql file beside it. (Node analogue: there's
// no clean one; it's like a bundler inlining a file as a string import, but the
// compiler does it natively.) The path is relative to THIS source file.
//
//go:embed migrations/0001_init.sql
var initSchema string

// Store wraps a connection POOL, not a single connection. A pool keeps a few
// live sockets ready and lends one out per query; opening a fresh TCP + auth
// handshake on every query would be brutally slow. (Node analogue: `pg.Pool`
// from node-postgres, or the pool Prisma/Drizzle manage for you.)
type Store struct {
	pool *pgxpool.Pool
}

// Connect dials the database and proves the link with a Ping.
//
// ctx (context.Context) is Go's cancellation + deadline handle, threaded through
// nearly every I/O call. If the caller gives up — request cancelled, timeout
// fires — the context is cancelled and the in-flight DB work unwinds with it.
// Mental model: like an AbortSignal you must pass explicitly to every async call
// instead of it being ambient.
func Connect(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping (is the db up on the right port?): %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close returns all pooled connections. Call it with `defer st.Close()`.
func (s *Store) Close() { s.pool.Close() }

// Migrate applies the embedded schema. Exec runs SQL that returns no rows
// (DDL like CREATE) — the counterpart to Query, which returns rows.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, initSchema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
