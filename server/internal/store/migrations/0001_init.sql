-- byakugan corpus schema. Idempotent: every statement is IF NOT EXISTS, so
-- running `corpus migrate` repeatedly is safe (no migration-version table yet —
-- one file, re-runnable, until the schema grows enough to need real versioning).

-- pgvector lives in the image but must be enabled per-database before the
-- `vector` column type exists. This is the line that turns plain Postgres into
-- a similarity-search engine.
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS chunks (
    id           TEXT PRIMARY KEY,        -- "RTA1987-s45A-en-0"
    authority    TEXT NOT NULL,           -- PDRM | JPJ | DBKL | RELIGIOUS
    statute      TEXT NOT NULL,
    statute_abbr TEXT NOT NULL,
    act_number   TEXT NOT NULL,           -- "333"
    state        TEXT NOT NULL,           -- ALL for federal
    section      TEXT NOT NULL,           -- "45A"
    heading      TEXT NOT NULL,
    lang         TEXT NOT NULL,           -- en | ms
    text         TEXT NOT NULL,           -- verbatim
    source_url   TEXT NOT NULL,
    as_at        TEXT,                    -- ISO date the text was current
    verified     BOOLEAN NOT NULL DEFAULT FALSE,
    embedding    vector(1024)             -- voyage-3 dim; NULL until `embed` runs
);

-- Cheap metadata filters retrieval will scope by (authority/state/lang) before
-- the vector search. A b-tree on these is plenty at this corpus size.
CREATE INDEX IF NOT EXISTS chunks_scope_idx ON chunks (authority, state, lang);
