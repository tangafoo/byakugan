// Package store is byakugan's door to the corpus warehouse: Postgres + pgvector.
// It hides SQL behind typed methods so the rest of the brain never hand-writes a
// query. Phase B slice 1: connect + migrate. Upsert/Search land in later slices.
package store

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"byakugan/internal/corpus"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// initSchema is the migration SQL compiled straight into the binary. //go:embed
// reads the file at BUILD time and assigns its contents to the string below — so
// the shipped binary needs no loose .sql file beside it. (Node analogue: there's
// no clean one; it's like a bundler inlining a file as a string import, but the
// compiler does it natively.) The path is relative to THIS source file.
//

//go:embed migrations/*.sql
var migrationFS embed.FS

const migrationsTableSQL = `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,	-- filename, e.g. "0002_identity.sql"
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

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

// Migrate applies the embedded schemas in order. applyMigration runs SQL that returns no
// rows (DDL like CREATE) — the counterpart to DML, which returns rows. Both
// files are idempotent (IF NOT EXISTS everywhere), so re-running is safe.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, migrationsTableSQL); err != nil {
		return fmt.Errorf("[migrate] could not set up version table: %w", err)
	}

	files, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("[migrate] trouble reading /migrations folder: %w", err)
	}

	for _, fullPath := range files {
		version := path.Base(fullPath)

		var applied bool
		if err := s.pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1);`,
			version).Scan(&applied); err != nil {
			return fmt.Errorf("[migrate] DB read error version %s: %w", version, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := migrationFS.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("[migrate] error reading file %s: %w", fullPath, err)
		}

		if err := s.applyMigration(ctx, string(sqlBytes), version); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) applyMigration(ctx context.Context, sql, version string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[migrate] error at begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("[migrate] could not migrate schema %s: %w", version, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1);`, version); err != nil {
		return fmt.Errorf("[migrate] error adding to version table: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[migrate] trouble committing: %w", err)
	}

	return nil
}

const upsertChunkSQL = `
	INSERT INTO chunks
		(id, authority, statute, statute_abbr, act_number, state, section, statute_code, subsection, kind, refs, heading, lang, text, source_url, as_at, verified, embedding)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	ON CONFLICT (id) DO UPDATE SET
		authority = EXCLUDED.authority,
		statute = EXCLUDED.statute,
		statute_abbr = EXCLUDED.statute_abbr,
		act_number = EXCLUDED.act_number,
		state = EXCLUDED.state,
		section = EXCLUDED.section,
		statute_code = EXCLUDED.statute_code,
		subsection = EXCLUDED.subsection,
		kind = EXCLUDED.kind,
		refs = EXCLUDED.refs,
		heading = EXCLUDED.heading,
		lang = EXCLUDED.lang,
		text = EXCLUDED.text,
		source_url = EXCLUDED.source_url,
		as_at = EXCLUDED.as_at,
		verified = EXCLUDED.verified,
		embedding = EXCLUDED.embedding;
`

func (s *Store) UpsertChunk(ctx context.Context, c corpus.Chunk, embedding []float32) error {
	refs, err := marshalRefs(c.Refs)
	if err != nil {
		return fmt.Errorf("[upsert] ID %s: %w", c.ID, err)
	}

	if _, err := s.pool.Exec(ctx, upsertChunkSQL, upsertArgs(c, refs, embedding)...); err != nil {
		return fmt.Errorf("error while upserting chunk to DB: %w", err)
	}
	return nil
}

func upsertArgs(c corpus.Chunk, refs []byte, embedding []float32) []any {
	return []any{
		c.ID,
		string(c.Authority),
		c.Statute,
		c.StatuteAbbr,
		c.ActNumber,
		string(c.State),
		c.Section,
		c.StatuteCode(),
		c.Subsection,
		string(c.Kind),
		refs,
		c.Heading,
		string(c.Lang),
		c.Text,
		c.SourceURL,
		c.AsAt,
		c.Verified,
		pgvector.NewVector(embedding),
	}
}

// marshalRefs encodes refs for the jsonb column; nil becomes the empty array
// (the column is NOT NULL DEFAULT '[]').
func marshalRefs(refs []corpus.RelatedSection) ([]byte, error) {
	if refs == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(refs)
	if err != nil {
		return nil, fmt.Errorf("[marshal] error turning refs into bytes: %w", err)
	}
	return b, nil
}

// ReplaceStatute/UpdateStatute swaps out one statute's chunks (in one language) atomically:
// delete everything under (statute_code, lang), reinsert the given chunks, one
// transaction. This is the re-slice safety net — when a section is re-cut into
// different slice IDs, plain upserts would leave the dead old IDs in the table
// with live embeddings, silently polluting retrieval.
func (s *Store) ReplaceStatute(ctx context.Context, statuteCode string, lang corpus.Lang, chunks []corpus.Chunk, embeddings [][]float32) error {
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("replace %s/%s: %d chunks but %d embeddings", statuteCode, lang, len(chunks), len(embeddings))
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("replace %s/%s: [begin]: %w", statuteCode, lang, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM chunks WHERE statute_code = $1 AND lang = $2;`,
		statuteCode, string(lang)); err != nil {
		return fmt.Errorf("replace %s/%s: delete: %w", statuteCode, lang, err)
	}

	for i, c := range chunks {
		refs, err := marshalRefs(c.Refs)
		if err != nil {
			return fmt.Errorf("refs to bytes %s/%s: %s: %w", statuteCode, lang, c.ID, err)
		}
		if _, err := tx.Exec(ctx, upsertChunkSQL, upsertArgs(c, refs, embeddings[i])...); err != nil {
			return fmt.Errorf("replace %s/%s: insert %s: %w", statuteCode, lang, c.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("replace %s/%s: commit: %w", statuteCode, lang, err)
	}
	return nil
}

// Different from Chunk because upserting a Chunk
// fills some nil values with fallbacks.
// Hits is like a polished ver of Chunk
type Hit struct {
	ID          string
	Section     string
	Heading     string
	Lang        string
	Text        string
	SourceURL   string
	Statute     string // "Dangerous Drugs Act 1952" — for citation display
	StatuteAbbr string // "DDA 1952"
	StatuteCode string // "DDA1952" — the matching key (abbr minus spaces)
	ActNumber   string
	Authority   string
	State       string
	Subsection  string // "12(2)-(4)" when the chunk is a slice; "" otherwise
	Kind        string
	Refs        []corpus.RelatedSection
	Verified    bool
	Distance    float64
}

// Same as corpus.Chunk.DisplaySection: the subsection
// span when sliced, the bare section otherwise.
func (h Hit) DisplaySection() string {
	if h.Subsection != "" {
		return h.Subsection
	}
	return h.Section
}

// refs comes back as jsonb bytes from the DB
const hitColumns = `id, section, heading, lang, text, source_url, statute, statute_abbr, statute_code, act_number,
	authority, state, subsection, kind, refs, verified`

const searchSQL = `
	SELECT ` + hitColumns + `, embedding <=> $1 AS distance
	FROM chunks
	WHERE embedding IS NOT NULL
		AND lang = $3
	ORDER BY distance
	LIMIT $2;
`

func scanHit(rows pgx.Rows) (Hit, error) {
	var h Hit
	var refs []byte
	if err := rows.Scan(&h.ID, &h.Section, &h.Heading, &h.Lang,
		&h.Text, &h.SourceURL, &h.Statute, &h.StatuteAbbr, &h.StatuteCode,
		&h.ActNumber, &h.Authority, &h.State, &h.Subsection, &h.Kind, &refs,
		&h.Verified, &h.Distance); err != nil {
		return Hit{}, fmt.Errorf("scan: %w", err)
	}
	// Unpackage refs from bytes to corpus.RelatedSection
	if err := json.Unmarshal(refs, &h.Refs); err != nil {
		return Hit{}, fmt.Errorf("scan refs of %s: %w", h.ID, err)
	}
	return h, nil
}

func (s *Store) Search(ctx context.Context, queryVec []float32, k int, l corpus.Lang, maxDist float64) ([]Hit, error) {
	rows, err := s.pool.Query(ctx, searchSQL,
		pgvector.NewVector(queryVec),
		k, l)
	if err != nil {
		return nil, fmt.Errorf("[search] trouble returning rows from chunks table: %w", err)
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		h, err := scanHit(rows)
		if err != nil {
			return nil, fmt.Errorf("search: %w", err)
		}
		hits = append(hits, h)
	}

	if maxDist <= 0 {
		return hits, rows.Err()
	}

	// Max distance filtering
	kept := hits[:0]

	for _, hit := range hits {
		if hit.Distance <= maxDist {
			kept = append(kept, hit)
		}
	}
	return kept, rows.Err()

}

// fetchSectionsSQL resolves a batch of (statute_code, section) pairs in one
// round trip. The unnest zip turns the two parallel arrays into a rowset —
// Postgres's way of doing WHERE (a,b) IN (list of tuples) with bind params.
// Distance is 0: these rows are graph edges (the statute's own cross-
// references), not similarity matches.
const fetchSectionsSQL = `
	SELECT ` + hitColumns + `, 0::float8 AS distance
	FROM chunks
	WHERE lang = $3
		AND (statute_code, section) IN (SELECT unnest($1::text[]), unnest($2::text[]))
	ORDER BY statute_code, section, id;
`

// FetchSections returns every chunk (all slices) matching any of the given
// refs, in the given language. Used for refs expansion: retrieval found via embedding search,
// refs bring out related sections for more accurate answers.
func (s *Store) FetchSections(ctx context.Context, refs []corpus.RelatedSection, lang corpus.Lang) ([]Hit, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	codes := make([]string, len(refs))
	sections := make([]string, len(refs))

	for i, r := range refs {
		codes[i] = r.Statute
		sections[i] = r.Section
	}

	rows, err := s.pool.Query(ctx, fetchSectionsSQL, codes, sections, string(lang))
	if err != nil {
		return nil, fmt.Errorf("fetch sections: %w", err)
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		h, err := scanHit(rows)
		if err != nil {
			return nil, fmt.Errorf("fetch sections: %w", err)
		}
		hits = append(hits, h)
	}

	return hits, rows.Err()
}
