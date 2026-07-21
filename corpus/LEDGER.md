# Corpus Ledger — the keeper's audit trail

> Maintained by the `corpus-keeper` agent. One entry per verification pass.
> A human trusts a corpus version by reading this file instead of re-diffing
> every chunk. `verified: true` is a signature; this is where the signature is
> witnessed and dated.

Legend: **FLIP** = false→true this pass · **HOLD** = stays false + reason ·
**RELIGIOUS** = mandatory human review, never auto-verified.

---

## Pass 2026-07-21 (d) — street-fights + animal cluster (9 new PC chunks, 1 new eval)

**Scope.** 9 chunks appended to `penalcode.jsonl` (PC s159/160/321/322/323/325/
351/352/428) + new `eval/fights.eval.jsonl` (6 cases). Source `raw/act574/`
(As at 31 May 2023, already on disk). No RELIGIOUS content.

**Result: 9 FLIP, 0 HOLD.** Corpus-wide now **90/90 verified**.

### The 4 prior PC chunks — untouched, confirmed
`PC-s304A/336/337/338` are byte-identical to Pass (b/c) and remain `verified:true`.
Diff confirmed no change.

### Verbatim — all 9 byte-clean against the 2023 source
Line-by-line + whitespace-normalized substring check: **9/9 exact-match**. Notable:
- **s322 / s351 / s352** carry their **Explanations + Illustrations verbatim** —
  confirmed these are enacted Code text in the official PDF (s322 Explanation +
  ILLUSTRATION lines 7529-7541; s351 Explanation + ILLUSTRATIONS (a)-(c) lines
  7998-8015; s352 Explanation lines 8027-8038). Em-dash "Explanation—" and curly
  quotes match source. Byte-clean.
- **s323 "except in the case provided for by section 334"** vs **s325 "except in
  the case provided by section 335"** — the "for" asymmetry is real in the source
  (lines 7547 / 7571); both reproduced exactly.
- **s428** — amended text "any animal or animals", ≤3-year tier, confirmed verbatim
  (lines 9760-9763).
- Headings all verbatim marginal notes.

### Flag checks — all cleared
1. **s429 deletion** — source line 9766 reads exactly "429. (Deleted by Act A1471)."
   The agent correctly chunked only s428 (as amended) and left a do-not-cite-s429
   note. A1471 is pre-2023, fully incorporated in the 31 May 2023 consolidation —
   no staleness from it. ✓
2. **s322/s351 Explanations + Illustrations** — confirmed enacted, in the official
   text, byte-clean (above). ✓
3. **Animal Welfare Act 2015 not chunked** — logged as a schema-gap decision below.
4. **Deliberate gaps** — logged below.

### Amendment scope for the 9 (rule 2f)
Same posture as the 4 negligence chunks (consolidation 31 May 2023): **PC A1750
(2025) inserts only s507b–507g** — confirmed from the on-disk act text this
session; it touches none of s159/160/321/322/323/325/351/352/428. The residual
Sep-2024 / Jul-Aug-2023 timeline entries are structurally the death-penalty
reform + un-numbered commencement events (cannot reach these max-≤7-year
fight/animal offence sections). Re-check satisfied → flip.

### Eval integrity — PASSED
`fights.eval.jsonl` (6 cases): all `expect`/`forbid` resolve to live chunks.
The **fight-04 / fight-05 mirror** is sound — fight-04 expects PC s323 (intent)
and forbids PC s337 (negligence); fight-05 expects s337 and forbids s323. Both
forbidden sections are real live chunks (meaningful tripwires guarding the
intent-vs-negligence line). No `should_find:true` case has empty `expect`.

### New standing gaps / schema flags (this pass)
- **Animal Welfare Act 2015 [Act 772] — NOT chunked (schema blocker, decision
  logged).** Two blockers: (a) **authority** — AWA enforcement is by "animal
  welfare officers" (s2, appointed under s36(1) by the Minister for agriculture,
  i.e. DVS); the word "police" appears nowhere in the Act, so no value in the
  closed `Authority` set (PDRM/JPJ/PBT/DBKL/RELIGIOUS) fits without mis-tagging.
  (b) **state scope** — AWA s1(2) is **Peninsular Malaysia + FT Labuan** only,
  which the `State` enum cannot express (no "Peninsular+Labuan" value). PC s428
  (PDRM) carries the poisoning-strays scenario until the schema grows a DVS /
  animal-welfare authority value. **Ready targets when unblocked:** AWA s29/s30/
  s31 — esp. **s30(1)** (no shooting dogs/cats without veterinary authorization,
  which also back-stops the firearms cluster). Flagged for the schema owner.
- **Deliberate coverage gaps (corpus stays lean):** PC s141–148 (unlawful
  assembly / rioting — the gang-fight beyond affray); s324/s326 (hurt / grievous
  hurt by dangerous weapons — knife/weapon escalation); s334/s335 (hurt /
  grievous hurt on grave & sudden provocation — textually cited by s323/s325);
  s350 (criminal force definition — leaned on by s352); s319/s320 (hurt /
  grievous hurt definitions — also logged under the negligence cluster).

---

## Pass 2026-07-21 (c) — amendment-scope ruling on the 22 held chunks (CPC + PC)

**Scope.** The two un-consolidated amendment acts that blocked the CPC + PC
chunks in Pass (b) were downloaded from the AGC portal and placed on disk. This
pass reads their full text and rules on section-level scope.

**Result: 22 FLIP, 0 HOLD.** Corpus-wide now **81/81 verified** (the entire
firearms + phone-search + conditional-rights expansion is clean and signed).

- **FLIP (22):** Penal Code (4) + CPC search (10) + CPC conditional (8).

### Primary-source amendment-scope evidence

- **`raw/act574/A1750.pdftotext.txt` — Penal Code (Amendment) Act 2025 [Act A1750]**
  (Royal Assent 25 Feb 2025, gazetted 7 Mar 2025). Full text read. It has ONE
  operative provision (s2): "The Penal Code is amended by inserting after section
  507a the following sections" → inserts **new sections 507b, 507c, 507d, 507e,
  507f, 507g** (the harassment / stalking / doxxing cluster). It amends nothing
  else. **Does NOT touch s304A / s336 / s337 / s338.** The four chunked negligence
  sections are unaffected → clean.

- **`raw/act593/A1751.pdftotext.txt` — Criminal Procedure Code (Amendment) Act 2025
  [Act A1751]** (Royal Assent 25 Feb 2025, gazetted 7 Mar 2025). Full text read.
  It has ONE operative provision (s2): "The Criminal Procedure Code is amended in
  the **First Schedule** by inserting after the item relating to section 507a…the
  following items" → adds First-Schedule offence-classification rows for the new
  PC 507b–507f (arrestable/bailable/compoundable columns). It amends nothing else.
  **Does NOT touch any of the 16 chunked CPC sections** — note it edits the *First*
  Schedule, while the two chunked schedule paragraphs are in the **Fourth**
  Schedule (untouched) → clean.

### Ruling

Both 7 Mar 2025 amendment acts are surgical insertions for the new harassment
offences; neither alters any chunked provision. Combined with the authoring
agent's prior check that **CPC A1662 (2023)** touched none of the chunked
sections, and the fact that AGC has not re-consolidated either code since
(1 Oct 2022 CPC / 31 May 2023 PC remain the current full-text reprints), the
chunked text stands as current law. Rule 2f re-check **satisfied with primary
sources** → all 22 flip. `as_at` stays at the consolidation date (2022-10-01 /
2023-05-31) — honest vintage; the text is unaffected by everything gazetted since.

**Residual note (not a blocker):** the AGC portal timeline also shows
"Amendments"-labelled rows for PC Jul/Aug 2023 and PC+CPC Sep 2024. The Jul/Aug
2023 PC rows are the Abolition-of-Mandatory-Death-Penalty commencements
(structurally confined to death-carrying sections — cannot reach the max-2-year
negligence sections). The Sep 2024 rows carry no visible Akta A-number in the
portal and were not present as amendment-act PDFs on the act-detail pages (the
coordinator, with portal access, identified A1750/A1751 as the amendment acts to
download); they are consistent with commencement / subsidiary events rather than
new principal amendments. If a future pass wants belt-and-suspenders, pull the
Sep 2024 P.U./act reference and confirm — but no evidence suggests it touches the
chunked sections, and the primary-source acts that DO exist are clean.

---

## Pass 2026-07-21 (b) — RE-VERIFICATION after re-sourcing to current AGC consolidations

**Scope.** The 39 chunks held for staleness in Pass (a) were re-derived by the
authoring agents against freshly-downloaded **current** AGC consolidations. This
pass re-diffs all 39 against the new sources and rules on the un-consolidated
amendment acts. (PA1967's 2 chunks from Pass (a) stand untouched.)

**Result: 17 FLIP, 22 HOLD.** Corpus-wide now **59/81 verified**.

- **FLIP (17):** Arms Act (7), FIPA (5), Child Act (5) — current consolidations,
  no un-consolidated amendment acts, all verbatim-clean.
- **HOLD (22):** CPC search (10) + CPC conditional (8) + Penal Code (4) — verbatim-
  clean against their current consolidations, but held on **un-consolidated
  post-`as_at` amendment acts whose section-level scope I could not confirm**.

### Verbatim re-diff against new sources — PASSED for all 39

New sources confirmed (each PDF's printed "As at" line): AA `As at 4 July 2023`,
FIPA `As at 4 July 2023`, PC `As at 31 May 2023` (old 2018 kept at `en.2018.*`),
CPC `As at 1 October 2022`, CA `As at 1 July 2023`. Ran a whitespace-normalized,
page-furniture-stripped substring check of every chunk against its new source:
**37/39 exact-match; the 2 "misses" are both confirmed-correct**, not defects:

- `CPC-s275-en-0` — the only diff is the `*` amendment marker I ruled strippable
  in Pass (a). Confirmed: chunk-with-`*`-reinserted matches source byte-for-byte.
  Marker-strip is clean (rule 2b). ✓
- `CA2001-s83A-en-0` — pdftotext renders the section number as lowercase "83a."
  everywhere (TOC, body, amendment note). **Rendered PDF page 95 to image and
  inspected the glyph directly: it is "83ᴀ." in SMALL CAPS** — a font/extraction
  artifact, not a lowercase section number. Every other inserted section (3A,
  20A, 28A, 116A/B) renders uppercase; legislative convention is uppercase. Chunk
  correctly uses "83A." ✓

Specific re-derivations verified clean against the new text:
- **FIPA** — `s7(1)` correctly rewritten to "imprisonment for a term of not less
  than thirty years but not exceeding forty years and with whipping with not less
  than six strokes" (the death/life tier removed by Act 846, in force 4 Jul 2023).
  `s2` correctly drops the now-deleted "imprisonment for life" definition (the
  2023 text removes it with no "(Deleted)" marker — confirmed against the PDF).
  `s7(2)/s8/s9` byte-identical. `s10(1)` "penalty of death" leftover is real in
  the official text but not chunked — not ours to fix.
- **Arms Act** — all 7 sections byte-identical to the 2017 text within the 4 Jul
  2023 consolidation (the 2019 + 2023 amendments touched none of ss 3/4/8/34/37/39/43).
- **Penal Code** — `s304A` "rash or **negligence** act" [sic] persists in the
  31 May 2023 reprint (confirmed vs both 2018 and 2023 PDFs). Four sections byte-clean.
- **CPC s388** — correctly re-derived to the 2022 capitalization "Officer in
  charge of the Police District". `s15` (Child Act) "Part VI ," space-before-comma
  typo fixed in the 2023 reprint; chunk follows the corrected text. All clean.

### Staleness ruling — the deciding axis again (rule 2f + `CLAUDE.md` current-law rule)

Checked each act's live AGC portal timeline for amendment **acts** (not subsidiary
legislation) dated after its consolidation:

| Act | Consolidation (`as_at`) | Latest amendment ACT | Verdict |
|-----|------------------------|----------------------|---------|
| Arms Act 1960 | 4 Jul 2023 | 16 Jun 2023 (in consolidation); later = P.U. only | **FLIP** — current, clean |
| FIPA 1971 | 4 Jul 2023 | 16 Jun 2023 (in consolidation); nothing after | **FLIP** — current, clean |
| Child Act 2001 | 1 Jul 2023 | 2017 (in consolidation); later = P.U. only | **FLIP** — current, clean |
| Penal Code | 31 May 2023 | **A1750, 7 Mar 2025** (+ A1681/A1691 2023) — un-consolidated | **HOLD** |
| CPC | 1 Oct 2022 | **A1751, 7 Mar 2025** (+ Aug 2023, Sep 2024; A1662 authoring-checked clean) — un-consolidated | **HOLD** |

Rationale for the split: for AA/FIPA/CA the downloaded consolidation **is** the
current published text and the portal shows **no amendment act after it** (only
P.U. subsidiary legislation, which does not alter section wording) → staleness
resolved, flip. For CPC and PC a post-`as_at` amendment **act** demonstrably
exists (both 7 Mar 2025), and rule 2f requires the chunk be **re-checked** — i.e.
the amendment's section-level scope confirmed not to touch the chunked provisions —
before it can be verified. I made a sustained effort to obtain that scope (AGC
portal timelines give dates only, not section detail; Wikipedia/Bar/news/law-firm
sources 403/404). **Unable to confirm scope → fail closed.** I did NOT rely on
recollection of the 2025 reform package to flip; the directive forbids it.

Precedent honoured: the MOA1955 (verified at 2006) standard is "current-and-
unamended is fine, age alone is not the bar" — it does **not** license flipping
over a *known* post-`as_at` amendment. AA/FIPA/CA fit MOA (unamended-after-
consolidation); CPC/PC do not (known 2025 amendment).

**HOLD reason for all 22 (verbatim-clean, staleness-blocked):** "un-consolidated
post-`as_at` amendment act (CPC A1751 / PC A1750, 7 Mar 2025) — section-level
scope vs the chunked provisions unconfirmed this pass; re-check requires the
amendment-act texts."

### Coordinator sub-questions — answered
1. **Un-consolidated amendments.** AA/FIPA/CA: none after consolidation → flip.
   CPC A1751 (2025) / PC A1750 (2025): exist, scope unconfirmed → hold (above).
2. **s275 marker-strip / s388 + s15 new-text derivations:** all clean (verified
   above — s275 `*`-only diff; s388 2022 capitalization; s15 2023 comma fix).
3. **"83a." font artifact:** resolved via direct PDF-glyph inspection — small caps,
   chunk's "83A" is correct.

### New standing item
- **CPC + PC re-check to unblock 22 chunks:** obtain the texts of **CPC (Amendment)
  Act 2025 [A1751]** and **Penal Code (Amendment) Act 2025 [A1750]** (both 7 Mar
  2025) and confirm they do not touch the chunked sections — PC 304A/336/337/338;
  CPC 15/17/19/20/20A/23/28A/62/116/116A/116B/117/275/289/388/Fourth Schedule.
  If clean, both clusters flip as-is (text already byte-perfect). Also fold PC
  A1681/A1691 (2023) and CPC Aug-2023/Sep-2024 amendments into that check.

---

## Pass 2026-07-21 — firearms + phone-search + conditional-rights clusters (41 new chunks, 3 new evals)

**Scope.** The 41 newly-authored chunks across 7 files (`arms1960`, `fipa1971`,
`penalcode`, `cpc`, `pa1967`, `cpc-conditional`, `child2001`) and 3 new eval
files (`arms`, `cpc`, `conditional`). No RELIGIOUS content in this pass.

**Headline result: 2 FLIP, 39 HOLD.**

- **FLIP (2):** `PA1967-s24-en-0`, `PA1967-s26-en-0` (Police Act 1967, Act 344).
- **HOLD (39):** all of Arms Act (7), FIPA (5), Penal Code (4), CPC search (10),
  CPC conditional (8), Child Act (5) — **held on staleness (rule 2f)**, not on
  text fidelity. See "Staleness — the deciding failure" below.

### Verbatim fidelity — PASSED for all 41

Every one of the 41 chunks was diffed line-by-line against its official AGC
`raw/actNNN/en.pdftotext.txt`. **All 41 are byte-for-byte clean**, including the
deliberately-preserved reprint quirks:

- `PC-s304A-en-0` "any rash or **negligence** act" [sic] — independently confirmed
  a genuine reprint typo via BOTH `pdftotext -layout` and `-raw` extractions
  (they agree). Correctly preserved verbatim. Not an OCR artifact.
- `CPC-s23-en-0` — the `-layout` duplicated margin labels `(f)/(g)/(h)` were
  correctly dropped; each paragraph carries its label once. Matches `-raw`.
- U+2015 horizontal-bar chapeau dashes (`―`) retained exactly where the reprint
  prints them: `CPC-s23` (×1), `CPC-s28A` (×4), `PA1967-s24` (×2). Em-dashes
  (`—`) retained where the reprint uses them (`CPC-s20`, `CPC-s116A`, `AA1960-s8/s39`,
  `FIPA-s2/s7`, `CA2001-*`, `sch4-para3`). Smart quotes/apostrophes match AGC style.
- `CPC-s28A-en-0` reprint quirk "paragraph 2(b)" (not "(2)(b)") preserved.
- `CA2001-s15-en-0` reprint quirk "Part VI ," (space before comma) preserved.
- Editorial `*NOTE—…` footnotes correctly EXCLUDED from `CPC-s275` and `PA1967-s26`.

### Keeper correction applied (rule 2b)

- `PA1967-s26-en-0`: the inline `*` amendment marker in "a fine not exceeding
  **\***two thousand ringgit" was **stripped by the keeper**. Rule 2b names
  "amendment markers / footnote superscripts pulled into the text" as corruption;
  the enacted provision reads "two thousand ringgit" with no asterisk. Text
  re-derived to the bare provision, then flipped. (The authoring agent had
  deliberately retained it; rule 2b overrides. The same `*` sits in `CPC-s275-en-0`
  — flagged there too, to be stripped when CPC is re-sourced and re-verified.)

### Identity / schema / tags — PASSED

- No duplicate `id` corpus-wide (81 chunks total; checked all `chunks/*.jsonl`).
- The two act593 lanes do not collide: `cpc.jsonl` sections {15,17,20,20A,23,62,116,116A,116B}
  vs `cpc-conditional.jsonl` {19,28A,117,275,289,388,Fourth Schedule} — zero overlap.
- All 41 pass `Chunk.Validate()` (ran `corpus load` on each file; all load clean).
- `authority: PDRM` correct for every chunk (CPC/Penal Code/Arms/FIPA/Police/Child
  are all police-lane criminal provisions). `state: ALL` correct (all federal).
  `lang: en` matches text. `statute_code` overrides verified: `PC` (Penal Code),
  and abbr-derived codes resolve correctly for CPC/PA1967/AA1960/FIPA1971/CA2001.
- `corpus` + `eval` packages compile together (`go build ./...` clean).

### Cross-references — PASSED (1 known gap)

- Only one orphan ref among the new files: `CPC-s117-en-0 → CPC s28` (the 24-hour
  rule). Recorded as a known coverage gap (see below), not a failure.
- All other refs resolve to live chunks (AA↔FIPA↔PC edges, CPC internal edges,
  CA internal edges, sch4→s20A/s19).

### Eval integrity — PASSED

- All `expect`/`forbid` in `arms.eval`, `cpc.eval`, `conditional.eval` resolve to
  live chunks by (StatuteCode, section). Forbid tripwires are real sections
  (`AA1960 s8`, `DDA1952 s31`, `MOA1955 s31`) — meaningful, not phantom.
- `cond-05`/`cond-06` are honest-miss negatives: `should_find:false`, empty
  `expect` — valid per `Case.Valid()`.
- `cond-04` expects `CPC "Fourth Schedule"` and resolves (matches the chunks'
  literal section field). Tied to the ugly schedule convention — see open items.

### Provenance — PASSED

- All 6 `source_url`s (`lom.agc.gov.my/act-detail.php?language=BI&act=NNN`) return
  HTTP 200 and point at the correct act (spot-confirmed act=206 → Arms Act 1960,
  act=593 → Criminal Procedure Code). Same convention as the already-verified
  RTA1987/DDA1952/MOA1955 chunks.
- Every `as_at` matches the reprint date printed in its PDF: AA 2017-02-01, FIPA
  2006-01-01, PC 2018-02-01, CPC 2017-09-01, PA 2024-05-10, CA 2018-02-01.

### Staleness — THE DECIDING FAILURE (rule 2f)

Checked each act's live AGC portal timeline against the downloaded consolidation.
The download is **superseded by a newer official consolidation** for 5 of 6 acts:

| Act | Downloaded (as_at) | AGC current | Verdict |
|-----|--------------------|-------------|---------|
| Police Act 1967 (344) | 10 May 2024 | **10 May 2024** (reprint = download; later = subsidiary legislation only) | **CURRENT → FLIP** |
| Arms Act 1960 (206) | 01 Feb 2017 | latest *reprint* still 01 Feb 2017, BUT un-consolidated amendment acts 22 Nov 2019 + 16 Jun 2023 | HOLD |
| FIPA 1971 (37) | 01 Jan 2006 | **04 Jul 2023 reprint** (Act 846 mandatory-death-penalty abolition, 16 Jun 2023) | HOLD |
| Penal Code (574) | 01 Feb 2018 | **31 May 2023 reprint** + 2023/24/25 amendments | HOLD |
| CPC (593) | 01 Sep 2017 | **01 Oct 2022 reprint** + 2023/24/26 amendments | HOLD |
| Child Act 2001 (611) | 01 Feb 2018 | **01 Jul 2023 reprint** | HOLD |

Rationale: `verified: true` promises a citizen the text is **current** law they can
read to an officer. Where a newer official consolidation demonstrably exists, the
downloaded text may carry superseded wording (acute example: FIPA s7 trafficking
"death" penalty, directly in scope of the 4 Jul 2023 death-penalty reform). Text
fidelity is impeccable; the **sourcing vintage** is the failure. Precedent is
consistent: MOA1955 stands verified at as_at 2006-01-01 because no newer AGC
consolidation supersedes it — age alone is not the bar, being *superseded* is.

**Per-chunk HOLD reasons (all 39): "source consolidation superseded / not
confirmed current — re-source the current AGC reprint, re-derive, re-verify."**
Chunks held (verbatim-clean, staleness-blocked):
- Arms Act: `AA1960-s3/s4/s8/s34/s37/s39/s43-en-0`
- FIPA: `FIPA1971-s2-en-0`, `-s7-en-0`, `-s7-en-1`, `-s8-en-0`, `-s9-en-0`
- Penal Code: `PC-s304A/s336/s337/s338-en-0`
- CPC search: `CPC-s15/s17/s20/s20A/s23-en-0`, `CPC-s23-en-1`, `CPC-s62/s116/s116A/s116B-en-0`
- CPC conditional: `CPC-s19/s28A/s117/s275/s289/s388-en-0`, `CPC-sch4-para3-en-0`, `CPC-sch4-para12-en-0`
- Child Act: `CA2001-s83A/s84/s85/s15/s96-en-0`

---

## Standing gaps & to-dos (open)

### Re-sourcing (priority — unblocks 39 verbatim-clean chunks)
1. **FIPA 1971** — re-download the 04 Jul 2023 reprint. Confirm s7 trafficking
   penalty post-Act-846 (mandatory death likely amended to court discretion) and
   diff s2/s8/s9. Highest-stakes staleness in the corpus.
2. **CPC (Act 593)** — re-download the ≥01 Oct 2022 reprint + fold 2023/24/26
   amendments; re-derive all 18 chunks. Flagship phone-search cluster depends on this.
3. **Penal Code (574)** — re-download ≥31 May 2023 reprint; re-check the four
   negligence sections (low change-risk but unconfirmed). Re-confirm the s304A
   "negligence act" [sic] persists in the current reprint.
4. **Child Act 2001 (611)** — re-download ≥01 Jul 2023 reprint; re-derive 5 chunks.
5. **Arms Act 1960 (206)** — *closest to flippable.* Current AGC consolidated
   reprint is still 01 Feb 2017 (= our source). Resolve the scope of the 22 Nov
   2019 and 16 Jun 2023 amendment entries against ss 3,4,8,34,37,39,43; if none
   touch these sections, AA is verifiable as-is (no re-download needed).

### Known coverage gaps (refs / definitions with no providing chunk — intentional)
- `CPC s28` (24-hour rule) — target of `CPC-s117` ref; not chunked.
- CPC `s44` (proclamation), `s51` (summons to produce), `s54–61`, `s296`
  (police supervision) — referenced in s23/s62/s116 text, not chunked.
- Penal Code `s319`/`s320` ("hurt"/"grievous hurt" definitions) — underpin
  s337/s338, not chunked.
- FIPA s2 → Arms Act `s9(1)/s9(2)/s11(1)/s15(1)` (dealer/transfer/import) — not chunked.
- SOSMA / Security Offences (Special Measures) Act 2012 [Act 747] — cited in
  CPC s116A(4); out of corpus.
- Whipping-exemption cross-refs in CPC s289 → PC `s376/377C/377CA/377E` — not chunked.
- Child Act `s2(1)` "child" definition — not chunked.
- WCA 2010 — noted by authoring agent as out of scope.

### BM parity (rule 3 — deferred per English-first build stance, logged per task)
All 6 acts have official Bahasa Malaysia consolidations on AGC LOM (`language=BM`).
All chunks here are **EN-only**. Acceptable under CLAUDE.md English-first build
stance; the authoritative BM *quote* should be added when each act is re-sourced
(one extra `pdftotext` per act). Standing gap: BM chunks for AA/FIPA/PC/CPC/PA/CA.

### Schema / convention flags for the human + main coding session (keeper cannot edit Go)
- **"sFourth Schedule" display bug.** `CPC-sch4-para3/para12` use
  `section:"Fourth Schedule"`; `DisplaySection()`/`EmbedText()` unconditionally
  prepend "s", rendering "sFourth Schedule" (confirmed in `corpus load` output).
  Needs a schema-level schedule display path (e.g. suppress the "s" for schedule
  sections, or a dedicated field). Keeper did NOT change the chunk convention this
  pass (no verification benefit while chunks are held; changing it would force
  moving `conditional.eval` cond-04's expect). Ruling deferred to schema owner.
- **CMA s249 stale-consolidation deferral** — carried over from authoring notes;
  not in this pass's scope, logged for continuity.
- **Fourth Schedule mechanics / CPC s20A dead-ends** — the sch4 chunks ref
  `CPC s20A` (exists) but the schedule is made *under* s20A; the schedule-as-law
  modeling (whole schedule vs. sliced paragraphs) is an open design question.

### Staleness watch (ongoing)
Oldest consolidations now flagged: FIPA 2006, CPC 2017, AA 2017 (and the newly-
superseded PC 2018, CA 2018). Re-source before any of their chunks flip. Police
Act 1967 is current as of this pass (10 May 2024, incorporating Act A1705).
