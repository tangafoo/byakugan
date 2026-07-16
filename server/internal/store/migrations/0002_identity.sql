-- 0002: statute identity + legal-unit metadata.
-- The corpus outgrew one statute; bare section numbers collide across acts
-- (DDA 1952 s31 vs MOA 1955 s31). These columns carry the identity that
-- retrieval, eval, and citations need — plus the refs cross-reference edges.
-- Same idempotent style as 0001: safe to run repeatedly, no version table.

ALTER TABLE chunks ADD COLUMN IF NOT EXISTS statute_code TEXT NOT NULL DEFAULT '';
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS subsection   TEXT NOT NULL DEFAULT '';
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS kind         TEXT NOT NULL DEFAULT '';
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS refs         JSONB NOT NULL DEFAULT '[]'::jsonb;

-- refs expansion + eval lookups resolve (statute_code, section, lang) directly.
CREATE INDEX IF NOT EXISTS chunks_section_lookup_idx ON chunks (statute_code, section, lang);
