# Corpus Rules — the law byakugan is allowed to speak

> This is the **source of truth** for corpus integrity. The `corpus-keeper` agent
> reads this file (and `CLAUDE.md`) before every verification pass. Humans edit
> this file to change the rules; the agent enforces whatever is written here.
>
> **The stakes, stated once.** A `verified: true` chunk is a promise: _this text
> is the law, word for word, and here is where to check it._ A citizen may read
> it aloud to an officer. `rasengan` may drop it into a real legal complaint. An
> external agent may consume it over MCP. Every one of those is a place a wrong
> chunk does real harm. There is no "close enough" in a verbatim quote. When in
> doubt, the chunk stays `false`. **Fail closed, always.**

---

## 0. The prime directive

**`verified` is a signature, not a checkbox.** The corpus-keeper is the only
actor permitted to flip a chunk from `verified: false` → `true`, and only after
**every** applicable check below passes against the official source. Any single
failure, ambiguity, missing source, or unresolvable doubt → the chunk stays
`false` and the reason is logged in the ledger. The keeper never flips `true` to
be "helpful," never on a hunch, never to unblock a build.

Corollary: **the keeper never paraphrases, "cleans up," or invents statute text
to make a check pass.** If `text` doesn't match source, the fix is to re-derive
the text _from the official raw source_ (below) — never to nudge it. If the
keeper can't re-derive cleanly, it flags for a human and leaves `verified: false`.

---

## 1. The layout it governs

| Path                              | What it is                                                            | Keeper's authority                                                              |
| --------------------------------- | --------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `corpus/chunks/*.jsonl`           | the retrievable law, one Chunk per line                               | read + edit (flip `verified`, fix refs/tags per rules, re-derive text from raw) |
| `corpus/eval/*.eval.jsonl`        | the golden gate, one Case per line                                    | read + edit (fix stale refs, flag orphans)                                      |
| `corpus/raw/actNNN/`              | **official source of truth** — PDFs + `*.pdftotext.txt`               | read only. This is the ground the keeper checks against.                        |
| `corpus/LEDGER.md`                | the keeper's own audit trail                                          | owns it — creates + maintains                                                   |
| `server/internal/corpus/chunk.go` | the schema (`Chunk`, `Authority`, `State`, `Lang`, `Kind`, `Valid()`) | read only — this defines the valid sets                                         |
| `server/internal/eval/case.go`    | the eval schema (`Case`, `SectionRef`)                                | read only                                                                       |

The keeper does **not** touch `server/` code or `app/`. Those belong to the human

- the main coding session. If a rule here requires a schema change, the keeper
  _flags it_ — it does not edit Go.

---

## 2. Per-chunk checklist

Run every applicable check against **the official source in `corpus/raw/actNNN/`**
(match `act_number` → folder). A chunk earns `verified: true` only when all pass.

### 2a. Identity & uniqueness

- **`id`** is present, stable, and follows the convention `<CODE><ActYear>-s<Section>[-<slice>]` (e.g. `RTA1987-s45A-0`, `DDA1952-s37-da`). One chunk, one id, forever.
- **No duplicate `id`** anywhere in `corpus/chunks/` (corpus-wide, not per-file).
- **`statute` / `statute_abbr` / `act_number` / `statute_code`** are mutually consistent and resolve to one real act. `StatuteCode()` (abbr minus spaces, or explicit `statute_code`) is the language-independent identity — a BM chunk displaying "APJ 1987" must still resolve to `RTA1987`.
- **`act_number`** matches the official Act No. and the `raw/actNNN/` folder it was sourced from (RTA 1987 = Act 333 = `raw/act333/`). Amendment acts are `A####`.

### 2b. The verbatim text — THE check

- **`text` is byte-for-byte the official statute text.** Compare against `raw/actNNN/*.pdftotext.txt` (and the PDF where the derived text is suspect). This is the single most important check; treat it as adversarial.
- **Common corruption to hunt** (all seen in this corpus already):
  - OCR / pdftotext drift: `1` vs `l`, `0` vs `O`, dropped diacritics, ligatures.
  - Page furniture bleeding in: running headers/footers, page numbers, "LAWS OF MALAYSIA", "Act 333".
  - **Shoulder / marginal notes** (the little side-headings) fused into body text.
  - Split-paragraph rejoin errors: a provision broken across a page glued back with a missing or doubled space.
  - Smart quotes vs straight quotes, en-dash vs em-dash vs hyphen, non-breaking spaces, trailing whitespace.
  - Amendment markers / footnote superscripts pulled into the text.
- **`heading`** is the real marginal note / section heading, verbatim — not a paraphrase and not invented.
- **The text is the _bare_ provision.** No embedding prefix, no framing, no editorial brackets. (`EmbedText()` adds context for retrieval; that is INPUT only and must never be what's stored as `text`.)

### 2c. Citation correctness

- **`section`** is the number a citizen reads aloud, exactly as the act numbers it ("45A", "37").
- **`subsection`** (when the section is sliced) accurately spans the provisions it claims ("12(2)-(4)", "37(da)") — the text present is _exactly_ those subsections, no more, no less.
- **`kind`** (optional) correctly labels the legal function (offence / power / procedure / presumption / penalty / defence / scope / definition). **Empty beats wrong** — an unsure `kind` is left blank, never guessed.

### 2d. Scoping tags — the app's whole routing depends on these

- **`authority`** is the body that actually enforces this provision, and is in the `Valid()` set (`PDRM` / `JPJ` / `PBT` / `DBKL` / `RELIGIOUS`). PDRM for CPC/Penal Code powers, JPJ for road-transport, PBT for LGA/SDBA local-authority powers, Religious for syariah. A mis-tagged authority sends a citizen the wrong answer at the wrong stop — treat as a hard fail.
- **`state`** is correct and in the `Valid()` set: `ALL` for federal acts (RTA, DDA, Penal Code), `PENINSULAR` where the act scopes itself so (MOA s1(2), LGA s1(1)), a specific state code for state enactments. Never `ALL` on a state-scoped syariah enactment.
- **`lang`** matches the actual language of `text` (`ms` for Bahasa, `en` for English). Mislabelling language silently breaks retrieval, which filters on it.

### 2e. Cross-references (`refs`)

- Every `refs[]` entry is a valid statute-qualified reference (`{statute, section}` both non-empty) and points at a **section that exists as a chunk** in the corpus — or the gap is recorded in the ledger as known-missing coverage.
- No self-reference (a section citing itself is always an authoring bug; `Validate()` catches exact matches, the keeper catches semantic ones too).
- `refs` statute codes resolve to real acts (`DDA1952`, not a typo).

### 2f. Provenance & staleness

- **`source_url`** is present, **resolves (HTTP 200)**, and points at _that act/provision_ on the official source (AGC / e-Federal Gazette) — not a search page, not a dead reprint, not a paywalled mirror.
- **`as_at`** is a valid ISO date (`YYYY-MM-DD`) reflecting when the text was current, and is **required** whenever `verified: true`. A word-for-word oath with no "as of when" is not an oath.
- The `as_at` is not obviously stale: if the source shows an amendment after `as_at`, the chunk is re-checked before it can stay `verified`.

### 2g. The gate

- **`verified`** flips to `true` only when 2a–2f all pass, and the keeper records in the ledger: the chunk id, the raw source file diffed against, the date, and "clean." Otherwise `false` + logged reason.

---

## 3. Corpus-wide checklist

Beyond per-chunk, the keeper sweeps the whole corpus:

- **No duplicate ids** across all `chunks/*.jsonl`.
- **No orphaned refs**: a `refs` target with no providing chunk is either filled or logged as a coverage gap (a gap is a to-do, not necessarily a failure — but a `verified` chunk leaning on a missing ref is a fragile citation and must be flagged).
- **Schema conformance**: every chunk still passes the current `Chunk.Validate()` in `chunk.go`. When `Valid()` sets change (a new authority/state/kind added), the keeper re-sweeps every chunk against the new sets.
- **Bilingual parity** (forward-looking): where an official Bahasa text exists for an act, the BM chunk is the legally authoritative one and should exist alongside the EN. Flag acts chunked EN-only when an official BM version is available. (BM is deferred for _framing_, not for the _quote_ — the quote should go bilingual as soon as both official texts exist.)

---

## 4. Eval integrity — the gate must be honest

The eval set is what proves retrieval works. A broken eval silently blesses broken
law. The keeper checks `corpus/eval/*.eval.jsonl`:

- Every `Case.expect[]` `SectionRef` points at a **section that exists in the corpus** (right statute code, right section). An eval expecting a section byakugan doesn't have is a false gate.
- Every `Case.forbid[]` `SectionRef` is meaningful — a tripwire for a section that _does_ exist (forbidding a nonexistent section proves nothing).
- `should_find: true` cases have non-empty `expect` (mirrors `Case.Valid()`).
- Statute codes in evals match live chunk `StatuteCode()`s (no `DDA1952` vs `DDA 1952` drift).
- **Code ↔ data ↔ eval compile together.** If `case.go` references a type (`SectionRef`) that `chunk.go` must provide, verify the corpus + eval packages actually build. A rename that lands in one file but not the other is a corpus-integrity bug the keeper flags. _(Known open item as of this writing: confirm `SectionRef` vs `RelatedSections` naming is consistent across `chunk.go` and `case.go`.)_

---

## 5. The high-stakes zone: religious / syariah content

Syariah enforcement content (`authority: RELIGIOUS`) gets the **highest scrutiny
and never auto-verifies**:

- It is **state-administered** — powers, offences, and procedures differ by state. `state` must be a specific state, never `ALL`, never `PENINSULAR`.
- It generally applies **only to Muslims** — applicability depends on state _and_ religion. Framing that overgeneralizes here is dangerous.
- It is legally nuanced, politically sensitive, and (per `CLAUDE.md`) should ideally see review by someone with legal expertise before it ships.
- **Rule: the keeper may run every check, but it flags syariah chunks for mandatory human legal review and leaves `verified: false` until a human signs off in the ledger.** The machine can catch OCR drift; it cannot judge whether a state enactment's scope is stated correctly. Fail closed, harder, here.

---

## 6. Two years out — invariants that must survive growth

Write and enforce as if byakugan is already:

- **Exposed over MCP** — external agents consume these chunks as ground truth. `verified: true` is a public API contract; a single byte-wrong verified chunk is a breach.
- **Feeding rasengan** — the action agent quotes only what the corpus hands it, so it _cannot_ hallucinate a section. That guarantee is only as true as this corpus. A wrong verified chunk becomes a wrong section in a real filed complaint.
- **Shipping an on-device index** — the offline bundle must match the server corpus exactly and version together. Flag any divergence between what's verified server-side and what's shipped.
- **Tracking amendments** — Malaysian law amends by amendment act (`A####`). When a source shows an amendment past a chunk's `as_at`, that chunk is stale until re-checked. The keeper watches for this and never lets a stale `verified` chunk stand silently.
- **Corpus-versioned** — the corpus as a whole carries a version; verified state is part of it. The ledger is the audit trail that makes a corpus version trustworthy.

---

## 7. The ledger (`corpus/LEDGER.md`)

The keeper's memory and audit trail. For each verification pass it records, per
chunk touched: id, verdict (verified / held-false), the raw source file diffed
against, `as_at` confirmed, date of check, and — for anything held `false` — the
exact reason. Open coverage gaps (missing refs, EN-only acts with BM available,
syariah pending human review) live here as a standing to-do. This file is how a
human trusts a corpus version without re-reading every chunk.

---

_Add rules below as byakugan teaches you what breaks. This doc is append-friendly
on purpose — every new failure mode the keeper (or a human) discovers becomes a
line here, and the keeper enforces it on the next pass._
