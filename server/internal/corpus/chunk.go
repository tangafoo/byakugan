// Package corpus is byakugan's brain-stem: the typed shape of Malaysian statute
// law as the app ingests, stores, and retrieves it. Every downstream phase —
// embeddings, pgvector retrieval, the Anthropic framing, the eval gate — speaks
// in these types. Get the shape right here and the rest has solid ground.
package corpus

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
	DBKL      Authority = "DBKL"      // KL City Hall — municipal by-laws, licensing, parking
	Religious Authority = "RELIGIOUS" // state syariah enforcement — state- + religion-scoped
)

func (a Authority) Valid() bool {
	switch a {
	case PDRM, JPJ, DBKL, Religious:
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
	case StateAll, Johor, Kedah, Kelantan, Melaka, NegeriSembilan, Pahang,
		Perak, Perlis, PulauPinang, Sabah, Sarawak, Selangor, Terengganu,
		KualaLumpur, Labuan, Putrajaya:
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

// DefaultAnswerLang is what we serve when the user hasn't chosen. Malay-first
// users facing a government setting are exactly who this app is for.
const DefaultAnswerLang = BM

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
	ID          string    `json:"id"`           // stable, e.g. "RTA1987-s45A-0"
	Authority   Authority `json:"authority"`    // who enforces this provision
	Statute     string    `json:"statute"`      // "Road Transport Act 1987"
	StatuteAbbr string    `json:"statute_abbr"` // "RTA 1987" — for compact display
	ActNumber   string    `json:"act_number"`   // "333" — the official Act No. (string: amendments are "A1234")
	State       State     `json:"state"`        // ALL for federal; specific otherwise
	Section     string    `json:"section"`      // "45A" — the number a citizen reads aloud
	Heading     string    `json:"heading"`      // marginal note / section heading
	Lang        Lang      `json:"lang"`         // language of Text (BM is authoritative)
	Text        string    `json:"text"`         // VERBATIM statute text. Never generated.
	SourceURL   string    `json:"source_url"`   // official source to point at / read from
	AsAt        string    `json:"as_at"`        // ISO date (YYYY-MM-DD) the text was current — for staleness detection
	Verified    bool      `json:"verified"`     // human-confirmed word-for-word?
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
	}
	return nil
}
