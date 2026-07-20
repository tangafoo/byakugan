# Layman

**Know your rights. Quote the law. Word for word.**

> Repo codename: `byakugan`. The product ships as **Layman**.

A know-your-rights app for Malaysians. When you're stopped — at a roadblock, at
your door, on the street — Layman shows you **exactly which section of the law
applies**, what the authority in front of you **can and cannot lawfully
require**, and what **your next move** is. Verbatim statute text, section
number, official source — something you can read out loud.

## The hard rules

1. **Verbatim law, never generated.** Every quote is *retrieved* word-for-word
   from the official statute text, with section number and source URL. The LLM
   frames and explains; it never writes the quote.
2. **Honest both ways.** When the authority genuinely has the power — to enter,
   search, arrest — Layman says so. Knowing when they're within their power is
   as protective as knowing when they're not.
3. **Never dead-end on a verdict.** Every answer carries three beats: the
   verdict, the limits that still bind the authority even while they act, and
   your next move — what to ask for, what to document, what recourse comes
   after.
4. **Which authority matters as much as which law.** PDRM, JPJ, local councils,
   religious enforcement — different bodies hold different statutory powers, and
   the same situation can have different answers depending on who's asking.
   Layman routes by authority (and state, and — for syariah matters — religion)
   before it answers.
5. **Fail soft.** No grounded section found → it says so plainly. It never
   improvises law.

## How it works

Grounded RAG over Malaysian statute law — the corpus is the statutes themselves,
not user content, so it's accurate on day one with zero users.

```
question ──► embed (Voyage) ──► pgvector search ──► rerank ──► distance floor
                                                                    │
        citations (verbatim, from Postgres — never touch the model) ─┤
                                                                    ▼
                                    Claude frames: verdict → limits → your move
```

The quote path and the model path never cross: statute text goes straight from
the database to the screen. The model only ever explains.

## Repo layout

| Path      | What                                                                |
| --------- | ------------------------------------------------------------------- |
| `server/` | Go backend — corpus CLI (`cmd/corpus`), Ask API (`cmd/api`), pgvector store, Voyage + Anthropic clients |
| `corpus/` | The law itself — official raw sources (`raw/`), chunked statutes (`chunks/`), eval golden set (`eval/`), `CORPUS_RULES.md` + ledger |
| `app/`    | Flutter app — Android / iOS / web, offline-first scenario cards (in progress) |

Corpus so far: Road Transport Act 1987, Dangerous Drugs Act 1952, Minor Offences
Act 1955, Local Government Act 1976, Street Drainage & Building Act 1974 —
bilingual (EN + BM) where official versions exist. Every chunk is tagged by
authority, state, and language; `verified: true` is earned by byte-for-byte
comparison against the official source (see `corpus/CORPUS_RULES.md`), never a
default.

## Run it

Needs Go, Docker, and API keys for [Voyage](https://voyageai.com) and
[Anthropic](https://anthropic.com) in `server/.env`.

```bash
docker compose up -d                                  # Postgres + pgvector (port 5433)
cd server
go run ./cmd/corpus migrate                           # schema
go run ./cmd/corpus embed ../corpus/chunks/rta1987.jsonl
go run ./cmd/corpus query "can police search my phone at a roadblock?"
go run ./cmd/corpus eval --rerank ../corpus/eval/rta1987.eval.jsonl   # the gate
```

The Ask API:

```bash
go run ./cmd/api    # listens on :8080
curl -s -X POST localhost:8080/ask -H 'Content-Type: application/json' \
  -d '{"question":"can the police search my phone at a roadblock?","lang":"en"}' | jq
```

## Status

Early and under construction. Retrieval + eval gate + online Ask are live;
Flutter shell and the offline on-device index are next.

## Not legal advice

Layman is an information tool. It surfaces the law and its official source; it
is not a lawyer and does not replace one. Laws amend — always check the cited
source, and get real legal help when it matters.
