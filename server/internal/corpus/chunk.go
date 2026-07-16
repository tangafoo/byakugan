// Package corpus is byakugan's brain-stem: the typed shape of Malaysian statute
// law as the app ingests, stores, and retrieves it. Every downstream phase —
// embeddings, pgvector retrieval, the Anthropic framing, the eval gate — speaks
// in these types. Get the shape right here and the rest has solid ground.
package corpus

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Authority — *who is in front of you*. Half the answer in this app. The same
// situation has different answers depending on the body, so retrieval is always
// scoped to one of these.
//
// TS analogue: `type Authority = "PDRM" | "JPJ" | ...` (a string-literal union).
// Go has no union types. The idiom is: a named string type + a fixed set of
// consts + a runtime validity check (see Valid below). You get the readable
// call sites of a union, but the compiler won't stop you assigning a bogus
// string — hence we validate at the boundary (on load), not in the type system.
// ─────────────────────────────────────────────────────────────────────────────

type Authority string

const (
	PDRM      Authority = "PDRM"      // Royal Malaysia Police — general criminal / CPC powers
	JPJ       Authority = "JPJ"       // Jabatan Pengangkutan Jalan — road transport
	PBT       Authority = "PBT"       // Pihak Berkuasa Tempatan — local authorities as a class (LGA 1976 / SDBA 1974 powers)
	DBKL      Authority = "DBKL"      // KL City Hall — the FT KL instance of PBT; for KL-specific by-laws
	Religious Authority = "RELIGIOUS" // state syariah enforcement — state- + religion-scoped
)

func (a Authority) Valid() bool {
	switch a {
	case PDRM, JPJ, PBT, DBKL, Religious:
		return true
	default:
		return false
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// State — powers vary by state for syariah and local-council law. Federal
// statutes (like the Road Transport Act 1987) apply nationwide → tag them All.
// ─────────────────────────────────────────────────────────────────────────────

type State string

const (
	StateAll State = "ALL" // federal statute — applies in every state

	// Peninsular Malaysia only — several federal acts scope themselves this way
	// (e.g. Minor Offences Act 1955 s1(2), Local Government Act 1976 s1(1));
	// Sabah/Sarawak have their own ordinances. Broader than one state, narrower
	// than ALL.
	Peninsular State = "PENINSULAR"

	Johor          State = "JHR"
	Kedah          State = "KDH"
	Kelantan       State = "KTN"
	Melaka         State = "MLK"
	NegeriSembilan State = "NSN"
	Pahang         State = "PHG"
	Perak          State = "PRK"
	Perlis         State = "PLS"
	PulauPinang    State = "PNG"
	Sabah          State = "SBH"
	Sarawak        State = "SWK"
	Selangor       State = "SGR"
	Terengganu     State = "TRG"
	KualaLumpur    State = "KUL" // Federal Territory
	Labuan         State = "LBN" // Federal Territory
	Putrajaya      State = "PJY" // Federal Territory
)

func (s State) Valid() bool {
	switch s {
	case StateAll, Peninsular, Johor, Kedah, Kelantan, Melaka, NegeriSembilan,
		Pahang, Perak, Perlis, PulauPinang, Sabah, Sarawak, Selangor,
		Terengganu, KualaLumpur, Labuan, Putrajaya:
		return true
	default:
		return false
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Lang — the language a piece of text is in. Two reasons this matters:
//   1. Malaysian statutes are authoritative in Bahasa Malaysia; the BM text is
//      the one you'd legally read to an officer. So a Chunk's Text carries its
//      own Lang (a chunk might be the BM original or an EN rendering).
//   2. The user picks the language of the *answer* (your request). That selector
//      rides on the Ask request and on each precomputed card variant — see
//      AnswerLang below. Retrieval stays multilingual regardless.
// ─────────────────────────────────────────────────────────────────────────────

type Lang string

const (
	BM Lang = "ms" // Bahasa Malaysia (ISO 639-1 "ms") — authoritative legal text
	EN Lang = "en" // English — secondary rendering / plain-language framing
)

func (l Lang) Valid() bool {
	switch l {
	case BM, EN:
		return true
	default:
		return false
	}
}

// DefaultAnswerLang is what we serve when the user hasn't chosen. English-first
// during the build (we optimize for accuracy, not language coverage); flips back
// to BM at the gov-sale milestone. See CLAUDE.md "Language & localization".
const DefaultAnswerLang = EN

// ─────────────────────────────────────────────────────────────────────────────
// RelatedSections — a pointer to one section of one statute, by statute code.
// The corpus now spans multiple acts, and bare section numbers collide (DDA 1952
// and MOA 1955 both have a s31) — so every cross-file reference must carry the
// statute too. Used two ways:
//   1. Chunk.Refs — statutory cross-references (s12 possession → s37 presumptions),
//      the edges retrieval follows to pull in the law's own footnotes.
//   2. eval.Case Expect/Forbid — statute-qualified expectations.
// ─────────────────────────────────────────────────────────────────────────────

type RelatedSection struct {
	Statute string `json:"statute"` // statute code = StatuteAbbr minus spaces, e.g. "DDA1952"
	Section string `json:"section"` // "37" — section level, not provision level
}

func (r RelatedSection) Valid() bool { return r.Statute != "" && r.Section != "" }

// ─────────────────────────────────────────────────────────────────────────────
// Kind — what a provision *does*, legally. A lawyer never reads a section flat:
// an offence, the power to arrest for it, the procedure that must be followed,
// and the presumption that shifts the burden are different tools. Tagging the
// function lets retrieval + framing compose the three-beat answer (verdict /
// limits / your move) from the right kinds of parts. Optional — leave empty
// rather than guess.
// ─────────────────────────────────────────────────────────────────────────────

type Kind string

const (
	Offence     Kind = "offence"     // creates the crime + usually its penalty
	Power       Kind = "power"       // grants an authority the power to act (arrest, enter, seize)
	Procedure   Kind = "procedure"   // how an act must lawfully be done (tests, notices)
	Presumption Kind = "presumption" // shifts the burden of proof onto the citizen
	Penalty     Kind = "penalty"     // penalty rules beyond the offence section itself
	Defence     Kind = "defence"     // statutory defences / escape hatches
	Scope       Kind = "scope"       // where/to whom the act applies
	Definition  Kind = "definition"  // interpretation sections
)

func (k Kind) Valid() bool {
	switch k {
	case Offence, Power, Procedure, Presumption,
		Penalty, Defence, Scope, Definition:
		return true
	default:
		return false
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Chunk — one retrievable unit of law: a section (or a slice of one) with the
// tags that let us scope retrieval and the provenance that lets us prove it.
//
// Hard rule made physical: Text is VERBATIM, retrieved — never generated. The
// Verified flag is the gate: until a human confirms Text is word-for-word from
// the official source, the chunk must not be quoted to a user.
//
// The `json:"..."` tags are struct tags — Go's reflection-time metadata. The
// encoding/json package reads them to map struct fields ↔ JSON keys. (Closest TS
// analogue: like a zod schema's key mapping, but it lives as a string annotation
// on the field rather than a separate schema object.)
// ─────────────────────────────────────────────────────────────────────────────

type Chunk struct {
	ID          string           `json:"id"`                     // stable, e.g. "RTA1987-s45A-0"
	Authority   Authority        `json:"authority"`              // who enforces this provision
	Statute     string           `json:"statute"`                // "Road Transport Act 1987"
	StatuteAbbr string           `json:"statute_abbr"`           // "RTA 1987" — for compact display
	ActNumber   string           `json:"act_number"`             // "333" — the official Act No. (string: amendments are "A1234")
	State       State            `json:"state"`                  // ALL for federal; specific otherwise
	Section     string           `json:"section"`                // "45A" — the number a citizen reads aloud
	ShortCode   string           `json:"statute_code,omitempty"` // canonical act code when it differs from the abbr-derived one — BM chunks display "APJ 1987" but ARE the same act as "RTA1987"
	Subsection  string           `json:"subsection,omitempty"`   // "12(2)-(4)" — the span within the section when it's sliced into legal units (covers paragraphs like "37(d)" too)
	Kind        Kind             `json:"kind,omitempty"`         // what the provision does (offence/power/presumption/...); optional
	Refs        []RelatedSection `json:"refs,omitempty"`         // statutory cross-references this provision leans on
	Heading     string           `json:"heading"`                // marginal note / section heading
	Lang        Lang             `json:"lang"`                   // language of Text (BM is authoritative)
	Text        string           `json:"text"`                   // VERBATIM statute text. Never generated.
	SourceURL   string           `json:"source_url"`             // official source to point at / read from
	AsAt        string           `json:"as_at"`                  // ISO date (YYYY-MM-DD) the text was current — for staleness detection
	Verified    bool             `json:"verified"`               // human-confirmed word-for-word?
}

// StatuteCode is the machine key for a statute — language-independent, since
// one act is one act whatever language its text is in. Most chunks omit
// statute_code in JSONL and derive it here (abbr minus spaces: "DDA 1952" →
// "DDA1952"); BM chunks set the ShortCode field explicitly because their
// display abbr is the Malay one ("APJ 1987") but their identity is still
// "RTA1987" — eval expectations and refs must match across languages.
// ALWAYS read the code through this getter, never the raw field.
func (c Chunk) StatuteCode() string {
	if c.ShortCode != "" {
		return c.ShortCode
	}
	return strings.ReplaceAll(c.StatuteAbbr, " ", "")
}

// DisplaySection is what a citation shows: the subsection span when this chunk
// is a slice ("37(da)"), the bare section otherwise ("37").
func (c Chunk) DisplaySection() string {
	if c.Subsection != "" {
		return c.Subsection
	}
	return c.Section
}

// EmbedText is the string sent to the embedding model — the verbatim text
// prefixed with its identity so retrieval can discriminate statutes and
// slices ("DDA 1952 s37(da) (presumption) — Presumptions\n<text>").
// HARD RULE GUARD: this is embedding INPUT only. What is stored, quoted, or
// framed as law is always the bare verbatim Text.
func (c Chunk) EmbedText() string {
	kind := ""
	if c.Kind != "" {
		kind = fmt.Sprintf(" (%s)", c.Kind)
	}
	return fmt.Sprintf("%s s%s%s — %s\n%s", c.StatuteAbbr, c.DisplaySection(), kind, c.Heading, c.Text)
}

// Validate enforces the boundary: a Chunk loaded from disk must have its tags in
// the known sets and the fields a citation can't live without. We check on the
// way IN (at load) so the rest of the pipeline can trust the data unconditionally.
//
// Go idiom note: returning an `error` value (not throwing). Callers handle it
// explicitly — the `if err != nil` you'll see everywhere. There are no exceptions
// in Go; errors are ordinary values you pass around and decide what to do with.
func (c Chunk) Validate() error {
	switch {
	case c.ID == "":
		return errMissing("id")
	case !c.Authority.Valid():
		return errBad("authority", string(c.Authority))
	case !c.State.Valid():
		return errBad("state", string(c.State))
	case !c.Lang.Valid():
		return errBad("lang", string(c.Lang))
	case c.Section == "":
		return errMissing("section")
	case c.Statute == "":
		return errMissing("statute")
	case c.SourceURL == "":
		return errMissing("source_url")
	case c.Text == "":
		return errMissing("text")
	case c.Verified && c.AsAt == "":
		// A chunk can't be sworn word-for-word true without saying *as of when*.
		// Pending chunks may omit as_at; verified ones may not.
		return errMissing("as_at (required when verified)")
	case c.Kind != "" && !c.Kind.Valid():
		return errBad("kind", string(c.Kind))
	}
	for i, r := range c.Refs {
		if !r.Valid() {
			return errBad(fmt.Sprintf("refs[%d]", i), r.Statute+":"+r.Section)
		}
		if r.Statute == c.StatuteCode() && r.Section == c.Section {
			// A section referencing itself is always an authoring mistake.
			return errBad(fmt.Sprintf("refs[%d] (self-reference)", i), r.Statute+":"+r.Section)
		}
	}
	return nil
}
