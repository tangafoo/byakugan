package corpus

import (
	"strings"
	"testing"
)

// goodChunk returns a minimal Chunk that passes Validate — each test case
// starts from this and breaks exactly one thing.
func goodChunk() Chunk {
	return Chunk{
		ID:          "DDA1952-s12-en-0",
		Authority:   PDRM,
		Statute:     "Dangerous Drugs Act 1952",
		StatuteAbbr: "DDA 1952",
		ActNumber:   "234",
		State:       StateAll,
		Section:     "12",
		Heading:     "Restriction on import and export of certain dangerous drugs",
		Lang:        EN,
		Text:        "12. (1) No person shall except under the authorization of the Minister—",
		SourceURL:   "https://lom.agc.gov.my/act-detail.php?language=BI&act=234",
		AsAt:        "2026-07-21",
		Verified:    false,
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Chunk)
		wantErr bool
	}{
		{"valid chunk", func(c *Chunk) {}, false},
		{"bad authority", func(c *Chunk) { c.Authority = "FBI" }, true},
		{"missing id", func(c *Chunk) { c.ID = "" }, true},
		{"bad state", func(c *Chunk) { c.State = "WAKANDA" }, true},
		{"bad lang", func(c *Chunk) { c.Lang = "fr" }, true},
		{"missing section", func(c *Chunk) { c.Section = "" }, true},
		{"missing statute", func(c *Chunk) { c.Statute = "" }, true},
		{"missing source_url", func(c *Chunk) { c.SourceURL = "" }, true},
		{"missing text", func(c *Chunk) { c.Text = "" }, true},
		{"verified without as_at", func(c *Chunk) { c.Verified = true; c.AsAt = "" }, true},
		{"verified with as_at", func(c *Chunk) { c.Verified = true }, false},
		{"pending may omit as_at", func(c *Chunk) { c.AsAt = "" }, false},
		{"bad kind", func(c *Chunk) { c.Kind = "vibe" }, true},
		{"empty kind is fine", func(c *Chunk) { c.Kind = "" }, false},
		{"ref missing section", func(c *Chunk) { c.Refs = []RelatedSection{{Statute: "DDA1952"}} }, true},
		{"self-reference", func(c *Chunk) { c.Refs = []RelatedSection{{Statute: "DDA1952", Section: "12"}} }, true},
		{"real ref is fine", func(c *Chunk) { c.Refs = []RelatedSection{{Statute: "DDA1952", Section: "37"}} }, false},
		// Self-reference is keyed on StatuteCode(), so a BM chunk with an
		// explicit ShortCode must still catch a ref to itself.
		{"self-reference via statute_code", func(c *Chunk) {
			c.StatuteAbbr = "APJ 1987"
			c.ShortCode = "DDA1952"
			c.Refs = []RelatedSection{{Statute: "DDA1952", Section: "12"}}
		}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chunk := goodChunk()
			tc.mutate(&chunk)
			err := chunk.Validate()

			if tc.wantErr && err == nil {
				t.Errorf("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected valid, got: %v", err)
			}
		})
	}
}

func TestStatuteCode(t *testing.T) {
	c := goodChunk()
	if got := c.StatuteCode(); got != "DDA1952" {
		t.Errorf("derived code: got %q, want %q", got, "DDA1952")
	}

	// BM chunk: Malay display abbr, explicit code override wins.
	c.StatuteAbbr = "APJ 1987"
	c.ShortCode = "RTA1987"
	if got := c.StatuteCode(); got != "RTA1987" {
		t.Errorf("shortcode override: got %q, want %q", got, "RTA1987")
	}
}

func TestDisplaySection(t *testing.T) {
	c := goodChunk()
	if got := c.DisplaySection(); got != "12" {
		t.Errorf("bare section: got %q, want %q", got, "12")
	}
	c.Subsection = "12(2)-(4)"
	if got := c.DisplaySection(); got != "12(2)-(4)" {
		t.Errorf("sliced section: got %q, want %q", got, "12(2)-(4)")
	}
}

// TestEmbedText pins the hard rule's shape: the embedding input carries the
// identity prefix, but the stored Text stays the bare provision — the prefix
// must never leak into what gets quoted as law.
func TestEmbedText(t *testing.T) {
	c := goodChunk()
	c.Kind = Offence

	got := c.EmbedText()
	for _, want := range []string{"DDA 1952", "s12", "(offence)", c.Heading, c.Text} {
		if !strings.Contains(got, want) {
			t.Errorf("EmbedText missing %q in:\n%s", want, got)
		}
	}

	if strings.Contains(c.Text, "DDA 1952 s12") {
		t.Errorf("identity prefix leaked into verbatim Text: %q", c.Text)
	}

	// No kind → no empty "()" residue.
	c.Kind = ""
	if strings.Contains(c.EmbedText(), "()") {
		t.Errorf("empty kind left residue in EmbedText: %q", c.EmbedText())
	}
}

func TestLoadJSONL(t *testing.T) {
	valid := `// a comment line
{"id":"MOA1955-s14-en-0","authority":"PDRM","statute":"Minor Offences Act 1955","statute_abbr":"MOA 1955","act_number":"336","state":"PENINSULAR","section":"14","heading":"Insulting behaviour","lang":"en","text":"14. Any person who uses any insulting behaviour...","source_url":"https://lom.agc.gov.my/act-detail.php?language=BI&act=336","as_at":"2026-07-21","verified":false}

{"id":"MOA1955-s21-en-0","authority":"PDRM","statute":"Minor Offences Act 1955","statute_abbr":"MOA 1955","act_number":"336","state":"PENINSULAR","section":"21","heading":"Drunkenness","lang":"en","text":"21. Any person found drunk and incapable...","source_url":"https://lom.agc.gov.my/act-detail.php?language=BI&act=336","as_at":"2026-07-21","verified":false}`

	chunks, err := LoadJSONL(strings.NewReader(valid))
	if err != nil {
		t.Fatalf("valid input errored: %v", err)
	}
	if len(chunks) != 2 {
		t.Errorf("got %d chunks, want 2 (comments and blanks must be skipped)", len(chunks))
	}

	t.Run("bad json reports line number", func(t *testing.T) {
		input := `{"id":"MOA1955-s14-en-0","authority":"PDRM","statute":"Minor Offences Act 1955","statute_abbr":"MOA 1955","act_number":"336","state":"PENINSULAR","section":"14","heading":"h","lang":"en","text":"t","source_url":"u","as_at":"2026-07-21","verified":false}
{not json at all`
		_, err := LoadJSONL(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("error should name line 2, got: %v", err)
		}
	})

	t.Run("invalid chunk fails the whole load", func(t *testing.T) {
		input := `{"id":"X-s1-en-0","authority":"FBI","statute":"X","statute_abbr":"X","act_number":"1","state":"ALL","section":"1","heading":"h","lang":"en","text":"t","source_url":"u","verified":false}`
		_, err := LoadJSONL(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile("testdata/does-not-exist.jsonl"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
