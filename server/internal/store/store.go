// Package store is byakugan's door to the corpus warehouse: Postgres + pgvector.
// It hides SQL behind typed methods so the rest of the brain never hand-writes a
// query. Phase B slice 1: connect + migrate. Upsert/Search land in later slices.
package store

import (
	"context"
	_ "embed"
	"fmt"

	"byakugan/internal/corpus"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
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
		return fmt.Errorf("could not migrate new schema: %w", err)
	}
	return nil
}

const upsertChunkSQL = `
	INSERT INTO chunks
		(id, authority, statute, statute_abbr, act_number, state, section, heading, lang, text, source_url, as_at, verified, embedding)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (id) DO UPDATE SET
		heading = EXCLUDED.heading,
		authority = EXCLUDED.authority,
		state = EXCLUDED.state,
		source_url = EXCLUDED.source_url,
		as_at = EXCLUDED.as_at,
		embedding = EXCLUDED.embedding,
		text = EXCLUDED.text,
		verified = EXCLUDED.verified;
`

func (s *Store) UpsertChunk(ctx context.Context, c corpus.Chunk, embedding []float32) error {
	if _, err := s.pool.Exec(ctx, upsertChunkSQL,
		c.ID,
		string(c.Authority),
		c.Statute,
		c.StatuteAbbr,
		c.ActNumber,
		string(c.State),
		c.Section,
		c.Heading,
		string(c.Lang),
		c.Text,
		c.SourceURL,
		c.AsAt,
		c.Verified,
		pgvector.NewVector(embedding)); err != nil {
		return fmt.Errorf("error while upserting chunk to DB: %w", err)
	}
	return nil
}

type Hit struct {
	ID        string
	Section   string
	Heading   string
	Lang      string
	Text      string
	Distance  float64
	SourceURL string
}

const searchSQL = `
	SELECT id, section, heading, lang, text, source_url, embedding <=> $1 AS distance
	FROM chunks
	WHERE embedding IS NOT NULL
		AND lang = $3
	ORDER BY distance
	LIMIT $2;
`

func (s *Store) Search(ctx context.Context, queryVec []float32, k int, l corpus.Lang) ([]Hit, error) {
	rows, err := s.pool.Query(ctx, searchSQL,
		pgvector.NewVector(queryVec),
		k, l)
	if err != nil {
		return nil, fmt.Errorf("[search] trouble returning rows from chunks table: %w", err)
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.ID, &h.Section, &h.Heading, &h.Lang, &h.Text, &h.SourceURL, &h.Distance); err != nil {
			return nil, fmt.Errorf("search: scan: %w", err)
		}
		hits = append(hits, h)
	}

	return hits, rows.Err()
}
